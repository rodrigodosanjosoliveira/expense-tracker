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

	"github.com/rodrigo/expense-tracker/internal/config"
	"github.com/rodrigo/expense-tracker/internal/database"
	"github.com/rodrigo/expense-tracker/internal/handler"
	"github.com/rodrigo/expense-tracker/internal/repository"
	"github.com/rodrigo/expense-tracker/internal/service"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/rodrigo/expense-tracker/docs" // Importa os docs gerados
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
	// Carregar configurações
	cfg := config.Load()

	// Conceito: Inicialização das dependências (Dependency Injection)
	var repo repository.ExpenseRepository
	var cleanup func()

	// Decidir qual repository usar baseado em variáveis de ambiente
	if os.Getenv("DB_HOST") != "" {
		// Usar PostgreSQL
		log.Println("Initializing PostgreSQL repository...")
		pool, err := database.NewPostgresPool(&cfg.Database)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		repo = repository.NewPostgresExpenseRepository(pool)
		cleanup = func() {
			log.Println("Closing database connection...")
			pool.Close()
		}
		log.Println("✓ Connected to PostgreSQL")
	} else {
		// Usar in-memory (desenvolvimento/testes)
		log.Println("Using in-memory repository (set DB_HOST to use PostgreSQL)")
		repo = repository.NewMemoryExpenseRepository()
		cleanup = func() {}
	}
	defer cleanup()

	idGen := service.NewUUIDGenerator()
	expenseService := service.NewExpenseService(repo, idGen)
	expenseHandler := handler.NewExpenseHandler(expenseService)

	// Conceito: Configurar rotas usando http.ServeMux (router padrão)
	mux := http.NewServeMux()

	// Rotas da API
	mux.HandleFunc("/expenses", func(w http.ResponseWriter, r *http.Request) {
		// Router simples baseado no método HTTP
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

	// Rota de health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Swagger UI - Documentação interativa da API
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// Conceito: Configurar servidor HTTP
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
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
