package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/rodrigo/expense-tracker/internal/config"
	"github.com/rodrigo/expense-tracker/internal/database"
	"github.com/rodrigo/expense-tracker/internal/handler"
	"github.com/rodrigo/expense-tracker/internal/middleware"
	"github.com/rodrigo/expense-tracker/internal/repository"
	"github.com/rodrigo/expense-tracker/internal/service"

	_ "github.com/rodrigo/expense-tracker/docs" // Importa os docs gerados
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Expense Tracker API
// @version 1.0
// @description API REST para controle de despesas pessoais
// @description Desenvolvida com Go usando TDD e arquitetura em camadas
//
// @contact.name Rodrigo
// @contact.email rodrigodosanjosoliveira@gmail.com
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /
// @schemes http

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Error validating config: %v", err)
	}

	// Conceito: Inicialização das dependências (Dependency Injection)
	var expenseRepo repository.ExpenseRepository
	var userRepo repository.UserRepository
	var webhookRepo repository.WebhookRepository
	var cleanup func()

	// Decidir qual repository usar baseado em variáveis de ambiente
	if cfg.UsePostgreSQL() {
		// Usar PostgreSQL
		log.Println("Initializing PostgreSQL repository...")
		pool, err := database.NewPostgresPool(&cfg.Database)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		expenseRepo = repository.NewPostgresExpenseRepository(pool)
		userRepo = repository.NewPostgresUserRepository(pool)
		webhookRepo = repository.NewPostgresWebhookRepository(pool)
		cleanup = func() {
			log.Println("Closing database connection...")
			pool.Close()
		}
		log.Println("✓ Connected to PostgreSQL")
	} else {
		// Usar in-memory (desenvolvimento/testes)
		log.Println("Using in-memory repository (set DB_HOST to use PostgreSQL)")
		expenseRepo = repository.NewMemoryExpenseRepository()
		// Note: UserRepository e WebhookRepository não têm implementação in-memory, PostgreSQL é obrigatório
		log.Println("⚠ WARNING: Authentication and webhooks require PostgreSQL. Set DB_HOST to enable.")
		cleanup = func() {}
	}

	defer cleanup()

	// Inicializar services
	idGen := service.NewUUIDGenerator()
	expenseService := service.NewExpenseService(expenseRepo, idGen)
	expenseHandler := handler.NewExpenseHandler(expenseService)

	// Inicializar autenticação (apenas se PostgreSQL estiver configurado)
	var authHandler *handler.AuthHandler
	var authMiddleware *middleware.AuthMiddleware

	// Inicializar webhooks e notificações (apenas se PostgreSQL estiver configurado)
	var webhookHandler *handler.WebhookHandler
	var notificationService *service.NotificationService

	if cfg.UsePostgreSQL() {
		// Autenticação
		authService := service.NewAuthService(userRepo, idGen, cfg.JWTSecret)
		authHandler = handler.NewAuthHandler(authService)
		authMiddleware = middleware.NewAuthMiddleware(authService)
		log.Println("✓ Authentication enabled")

		// Webhooks e notificações
		notificationService = service.NewNotificationService(webhookRepo)
		webhookService := service.NewWebhookService(webhookRepo, idGen)
		webhookHandler = handler.NewWebhookHandler(webhookService)

		// Injetar notification service no expense service para disparar notificações
		expenseService.SetNotificationService(notificationService)

		log.Println("✓ Webhooks and notifications enabled")
	} else {
		log.Println("⚠ Authentication and webhooks disabled (PostgreSQL required)")
	}

	// Conceito: Configurar rotas usando http.ServeMux (router padrão)
	mux := http.NewServeMux()

	// Rotas de autenticação (públicas)
	if authHandler != nil {
		mux.HandleFunc("/auth/register", authHandler.Register)
		mux.HandleFunc("/auth/login", authHandler.Login)
		log.Println("✓ Auth routes registered: /auth/register, /auth/login")
	}

	// Rotas da API de expenses (protegidas com autenticação se disponível)
	if authMiddleware != nil {
		// Com autenticação
		mux.HandleFunc("/expenses", func(w http.ResponseWriter, r *http.Request) {
			handler := authMiddleware.Authenticate(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					if r.URL.Path == "/expenses" || r.URL.Path == "/expenses/" {
						expenseHandler.ListExpenses(w, r)
					} else {
						expenseHandler.GetExpense(w, r)
					}
				case http.MethodPost:
					expenseHandler.CreateExpense(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			})
			handler(w, r)
		})

		mux.HandleFunc("/expenses/", func(w http.ResponseWriter, r *http.Request) {
			handler := authMiddleware.Authenticate(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					expenseHandler.GetExpense(w, r)
				case http.MethodPut:
					expenseHandler.UpdateExpense(w, r)
				case http.MethodDelete:
					expenseHandler.DeleteExpense(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			})
			handler(w, r)
		})
		log.Println("✓ Expense routes protected with authentication")
	} else {
		// Sem autenticação (modo desenvolvimento)
		mux.HandleFunc("/expenses", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				if r.URL.Path == "/expenses" || r.URL.Path == "/expenses/" {
					expenseHandler.ListExpenses(w, r)
				} else {
					expenseHandler.GetExpense(w, r)
				}
			case http.MethodPost:
				expenseHandler.CreateExpense(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})

		mux.HandleFunc("/expenses/", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				expenseHandler.GetExpense(w, r)
			case http.MethodPut:
				expenseHandler.UpdateExpense(w, r)
			case http.MethodDelete:
				expenseHandler.DeleteExpense(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})
		log.Println("⚠ Expense routes running WITHOUT authentication")
	}

	// Rotas de webhooks (protegidas com autenticação)
	if webhookHandler != nil && authMiddleware != nil {
		mux.HandleFunc("/webhooks", func(w http.ResponseWriter, r *http.Request) {
			handler := authMiddleware.Authenticate(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					webhookHandler.ListWebhooks(w, r)
				case http.MethodPost:
					webhookHandler.CreateWebhook(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			})
			handler(w, r)
		})

		mux.HandleFunc("/webhooks/", func(w http.ResponseWriter, r *http.Request) {
			handler := authMiddleware.Authenticate(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					webhookHandler.GetWebhook(w, r)
				case http.MethodPut:
					webhookHandler.UpdateWebhook(w, r)
				case http.MethodDelete:
					webhookHandler.DeleteWebhook(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			})
			handler(w, r)
		})
		log.Println("✓ Webhook routes registered: /webhooks")
	}

	// Rota de health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Swagger UI - Documentação interativa da API
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// Conceito: Configurar servidor HTTP
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Conceito: Graceful shutdown - desligar servidor graciosamente
	// Canal para receber sinais do sistema operacional
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Conceito: Goroutine - executar servidor em background
	go func() {
		log.Printf("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Aguardar sinal de interrupção
	<-stop
	log.Println("Shutting down server...")

	// Conceito: Context com timeout para graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}
