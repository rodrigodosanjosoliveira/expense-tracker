package handler

// Tests for CategoryHandler covering acceptance criteria:
//   AC-1: CRUD de categorias com user_id scoping
//   AC-2: Unicidade case-insensitive de nome por usuario (HTTP 409)
//   AC-7: DELETE categoria em uso retorna HTTP 409

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rodrigo/expense-tracker/internal/domain"
	"github.com/rodrigo/expense-tracker/internal/middleware"
	"github.com/rodrigo/expense-tracker/internal/repository"
	"github.com/rodrigo/expense-tracker/internal/service"
)

const (
	categoriesPath    = "/categories"
	categoryByIDPath  = "/categories/cat-id-123"
	testCategoryID    = "cat-id-123"
	otherUserID       = "other-user-id"
)

// mockCategoryIDGen returns a deterministic category ID for tests.
type mockCategoryIDGen struct{}

func (m *mockCategoryIDGen) Generate() string {
	return testCategoryID
}

// setupCategoryHandler creates a CategoryHandler backed by in-memory repositories.
// CategoryHandler and CategoryService are not yet implemented — tests will fail
// to compile until Dev implements them (RED phase).
func setupCategoryHandler() *CategoryHandler {
	repo := repository.NewMemoryCategoryRepository()
	idGen := &mockCategoryIDGen{}
	svc := service.NewCategoryService(repo, idGen)
	return NewCategoryHandler(svc)
}

func withCategoryUserID(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
	return r.WithContext(ctx)
}

// ---------------------------------------------------------------------------
// AC-1: POST /categories — criar categoria
// ---------------------------------------------------------------------------

func TestCategoryHandlerCreateCategory(t *testing.T) {
	tests := []struct {
		name       string
		body       interface{}
		addUserID  bool
		wantStatus int
	}{
		{
			name:       "valid category",
			body:       map[string]string{"name": "Alimentacao"},
			addUserID:  true,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "empty name returns 400",
			body:       map[string]string{"name": ""},
			addUserID:  true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "name too long returns 400",
			body:       map[string]string{"name": string(make([]byte, 101))},
			addUserID:  true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid JSON returns 400",
			body:       "invalid-json",
			addUserID:  true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing JWT returns 401",
			body:       map[string]string{"name": "Transporte"},
			addUserID:  false,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setupCategoryHandler()

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, categoriesPath, bytes.NewReader(body))
			if tt.addUserID {
				req = withCategoryUserID(req, testUserID)
			}
			w := httptest.NewRecorder()

			h.CreateCategory(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateCategory() status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantStatus == http.StatusCreated {
				var result domain.Category
				if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
					t.Fatalf("CreateCategory() failed to decode response: %v", err)
				}
				if result.ID == "" {
					t.Error("CreateCategory() returned empty ID")
				}
				if result.UserID != testUserID {
					t.Errorf("CreateCategory() UserID = %q, want %q", result.UserID, testUserID)
				}
				if result.Name != "Alimentacao" {
					t.Errorf("CreateCategory() Name = %q, want %q", result.Name, "Alimentacao")
				}
			}
		})
	}
}

// AC-1: response must include all required fields
func TestCategoryHandlerCreateCategoryResponseShape(t *testing.T) {
	h := setupCategoryHandler()

	body, _ := json.Marshal(map[string]string{"name": "Lazer"})
	req := withCategoryUserID(httptest.NewRequest(http.MethodPost, categoriesPath, bytes.NewReader(body)), testUserID)
	w := httptest.NewRecorder()

	h.CreateCategory(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("CreateCategory() status = %d, want %d", w.Code, http.StatusCreated)
	}

	var result domain.Category
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if result.CreatedAt.IsZero() {
		t.Error("CreateCategory() CreatedAt is zero")
	}
	if result.UpdatedAt.IsZero() {
		t.Error("CreateCategory() UpdatedAt is zero")
	}
}

// ---------------------------------------------------------------------------
// AC-2: Unicidade case-insensitive — HTTP 409 on duplicate
// ---------------------------------------------------------------------------

func TestCategoryHandlerCreateDuplicateName(t *testing.T) {
	h := setupCategoryHandler()

	// Create first category
	body, _ := json.Marshal(map[string]string{"name": "Alimentacao"})
	req1 := withCategoryUserID(httptest.NewRequest(http.MethodPost, categoriesPath, bytes.NewReader(body)), testUserID)
	w1 := httptest.NewRecorder()
	h.CreateCategory(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("first CreateCategory failed with status %d", w1.Code)
	}

	tests := []struct {
		name       string
		inputName  string
		wantStatus int
	}{
		{
			name:       "exact duplicate returns 409",
			inputName:  "Alimentacao",
			wantStatus: http.StatusConflict,
		},
		{
			name:       "case-insensitive duplicate returns 409",
			inputName:  "alimentacao",
			wantStatus: http.StatusConflict,
		},
		{
			name:       "mixed-case duplicate returns 409",
			inputName:  "ALIMENTACAO",
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dupBody, _ := json.Marshal(map[string]string{"name": tt.inputName})
			req := withCategoryUserID(httptest.NewRequest(http.MethodPost, categoriesPath, bytes.NewReader(dupBody)), testUserID)
			w := httptest.NewRecorder()
			h.CreateCategory(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateCategory() duplicate %q: status = %d, want %d", tt.inputName, w.Code, tt.wantStatus)
			}
		})
	}
}

// AC-2: same name by different user must succeed
func TestCategoryHandlerSameNameDifferentUser(t *testing.T) {
	h := setupCategoryHandler()

	body, _ := json.Marshal(map[string]string{"name": "Alimentacao"})

	// User A creates category
	req1 := withCategoryUserID(httptest.NewRequest(http.MethodPost, categoriesPath, bytes.NewReader(body)), testUserID)
	w1 := httptest.NewRecorder()
	h.CreateCategory(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("user A CreateCategory failed with status %d", w1.Code)
	}

	// User B creates same name — must succeed (201)
	body2, _ := json.Marshal(map[string]string{"name": "Alimentacao"})
	req2 := withCategoryUserID(httptest.NewRequest(http.MethodPost, categoriesPath, bytes.NewReader(body2)), otherUserID)
	w2 := httptest.NewRecorder()
	h.CreateCategory(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Errorf("user B CreateCategory same name: status = %d, want %d (users must be isolated)", w2.Code, http.StatusCreated)
	}
}

// ---------------------------------------------------------------------------
// AC-1: GET /categories — list (user_id scoped)
// ---------------------------------------------------------------------------

func TestCategoryHandlerListCategories(t *testing.T) {
	h := setupCategoryHandler()

	// Seed: user A has 2 categories, user B has 1
	seedCategory(t, h, "Alimentacao", testUserID)
	seedCategory(t, h, "Transporte", testUserID)
	seedCategory(t, h, "Saude", otherUserID)

	req := withCategoryUserID(httptest.NewRequest(http.MethodGet, categoriesPath, nil), testUserID)
	w := httptest.NewRecorder()
	h.ListCategories(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ListCategories() status = %d, want %d", w.Code, http.StatusOK)
	}

	var result []*domain.Category
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("ListCategories() decode failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("ListCategories() returned %d categories, want 2 (user_id scoping broken)", len(result))
	}

	for _, cat := range result {
		if cat.UserID != testUserID {
			t.Errorf("ListCategories() returned category with UserID = %q, want %q", cat.UserID, testUserID)
		}
	}
}

func TestCategoryHandlerListCategoriesUnauthorized(t *testing.T) {
	h := setupCategoryHandler()

	req := httptest.NewRequest(http.MethodGet, categoriesPath, nil) // no user_id in context
	w := httptest.NewRecorder()
	h.ListCategories(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("ListCategories() without auth: status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ---------------------------------------------------------------------------
// AC-1: GET /categories/{id}
// ---------------------------------------------------------------------------

func TestCategoryHandlerGetCategory(t *testing.T) {
	h := setupCategoryHandler()
	seedCategory(t, h, "Alimentacao", testUserID)

	tests := []struct {
		name       string
		path       string
		userID     string
		wantStatus int
	}{
		{
			name:       "owner gets own category",
			path:       "/categories/" + testCategoryID,
			userID:     testUserID,
			wantStatus: http.StatusOK,
		},
		{
			name:       "other user gets 404 (not 403 — must not reveal existence)",
			path:       "/categories/" + testCategoryID,
			userID:     otherUserID,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "non-existent ID returns 404",
			path:       "/categories/does-not-exist",
			userID:     testUserID,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "empty ID returns 400",
			path:       "/categories/",
			userID:     testUserID,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := withCategoryUserID(httptest.NewRequest(http.MethodGet, tt.path, nil), tt.userID)
			w := httptest.NewRecorder()
			h.GetCategory(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("GetCategory() %s: status = %d, want %d", tt.name, w.Code, tt.wantStatus)
			}
		})
	}
}

func TestCategoryHandlerGetCategoryUnauthorized(t *testing.T) {
	h := setupCategoryHandler()

	req := httptest.NewRequest(http.MethodGet, categoryByIDPath, nil)
	w := httptest.NewRecorder()
	h.GetCategory(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetCategory() without auth: status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ---------------------------------------------------------------------------
// AC-1: PUT /categories/{id}
// ---------------------------------------------------------------------------

func TestCategoryHandlerUpdateCategory(t *testing.T) {
	h := setupCategoryHandler()
	seedCategory(t, h, "Alimentacao", testUserID)

	updateBody, _ := json.Marshal(map[string]string{"name": "Alimentacao Saudavel"})

	tests := []struct {
		name       string
		path       string
		userID     string
		body       []byte
		wantStatus int
	}{
		{
			name:       "owner updates own category",
			path:       "/categories/" + testCategoryID,
			userID:     testUserID,
			body:       updateBody,
			wantStatus: http.StatusOK,
		},
		{
			name:       "other user cannot update — returns 404",
			path:       "/categories/" + testCategoryID,
			userID:     otherUserID,
			body:       updateBody,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "empty name returns 400",
			path:       "/categories/" + testCategoryID,
			userID:     testUserID,
			body:       func() []byte { b, _ := json.Marshal(map[string]string{"name": ""}); return b }(),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid JSON returns 400",
			path:       "/categories/" + testCategoryID,
			userID:     testUserID,
			body:       []byte("not-json"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "non-existent ID returns 404",
			path:       "/categories/does-not-exist",
			userID:     testUserID,
			body:       updateBody,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := withCategoryUserID(httptest.NewRequest(http.MethodPut, tt.path, bytes.NewReader(tt.body)), tt.userID)
			w := httptest.NewRecorder()
			h.UpdateCategory(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("UpdateCategory() %s: status = %d, want %d (body: %s)", tt.name, w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestCategoryHandlerUpdateCategoryUnauthorized(t *testing.T) {
	h := setupCategoryHandler()
	body, _ := json.Marshal(map[string]string{"name": "Novo Nome"})
	req := httptest.NewRequest(http.MethodPut, categoryByIDPath, bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.UpdateCategory(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("UpdateCategory() without auth: status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// AC-2: update to name that already exists for same user returns 409
func TestCategoryHandlerUpdateDuplicateName(t *testing.T) {
	h := setupCategoryHandler()
	seedCategory(t, h, "Alimentacao", testUserID)

	// Create a second category with a different ID
	secondRepo := repository.NewMemoryCategoryRepository()
	secondIDGen := &fixedIDGen{id: "cat-id-456"}
	secondSvc := service.NewCategoryService(secondRepo, secondIDGen)
	h2 := NewCategoryHandler(secondSvc)
	// We need both categories in the same handler/service, so seed them manually
	// via the same handler with sequential IDs.
	// This test uses a separate handler with a controlled IDGen for the second ID.
	_ = h2

	// Simpler approach: seed both categories on a fresh handler with sequential IDs.
	hFresh := setupCategoryHandlerWithIDs("cat-id-001", "cat-id-002")

	seedCategoryOnHandler(t, hFresh, "Alimentacao", testUserID, "cat-id-001")
	seedCategoryOnHandler(t, hFresh, "Transporte", testUserID, "cat-id-002")

	// Try to rename "Transporte" to "alimentacao" (case-insensitive collision with first)
	updateBody, _ := json.Marshal(map[string]string{"name": "alimentacao"})
	req := withCategoryUserID(httptest.NewRequest(http.MethodPut, "/categories/cat-id-002", bytes.NewReader(updateBody)), testUserID)
	w := httptest.NewRecorder()
	hFresh.UpdateCategory(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("UpdateCategory() name collision: status = %d, want %d", w.Code, http.StatusConflict)
	}
}

// ---------------------------------------------------------------------------
// AC-1 + AC-7: DELETE /categories/{id}
// ---------------------------------------------------------------------------

func TestCategoryHandlerDeleteCategory(t *testing.T) {
	h := setupCategoryHandler()
	seedCategory(t, h, "Alimentacao", testUserID)

	tests := []struct {
		name       string
		path       string
		userID     string
		wantStatus int
	}{
		{
			name:       "owner deletes own category",
			path:       "/categories/" + testCategoryID,
			userID:     testUserID,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "non-existent after deletion returns 404",
			path:       "/categories/" + testCategoryID,
			userID:     testUserID,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := withCategoryUserID(httptest.NewRequest(http.MethodDelete, tt.path, nil), tt.userID)
			w := httptest.NewRecorder()
			h.DeleteCategory(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("DeleteCategory() %s: status = %d, want %d", tt.name, w.Code, tt.wantStatus)
			}
		})
	}
}

func TestCategoryHandlerDeleteOtherUserCategory(t *testing.T) {
	h := setupCategoryHandler()
	seedCategory(t, h, "Alimentacao", testUserID)

	req := withCategoryUserID(httptest.NewRequest(http.MethodDelete, "/categories/"+testCategoryID, nil), otherUserID)
	w := httptest.NewRecorder()
	h.DeleteCategory(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("DeleteCategory() other user: status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestCategoryHandlerDeleteCategoryUnauthorized(t *testing.T) {
	h := setupCategoryHandler()

	req := httptest.NewRequest(http.MethodDelete, categoryByIDPath, nil)
	w := httptest.NewRecorder()
	h.DeleteCategory(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("DeleteCategory() without auth: status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// AC-7: DELETE category in use returns HTTP 409
func TestCategoryHandlerDeleteCategoryInUse(t *testing.T) {
	repo := repository.NewMemoryCategoryRepository()
	idGen := &mockCategoryIDGen{}
	svc := service.NewCategoryService(repo, idGen)
	h := NewCategoryHandler(svc)

	// Seed the category
	seedCategoryViaRepo(repo, testCategoryID, "Alimentacao", testUserID)

	// Mark it as in-use
	repo.MarkCategoryInUse(testCategoryID)

	req := withCategoryUserID(httptest.NewRequest(http.MethodDelete, "/categories/"+testCategoryID, nil), testUserID)
	w := httptest.NewRecorder()
	h.DeleteCategory(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("DeleteCategory() in-use: status = %d, want %d", w.Code, http.StatusConflict)
	}
}

// ---------------------------------------------------------------------------
// AC-1: Method Not Allowed
// ---------------------------------------------------------------------------

func TestCategoryHandlerMethodNotAllowed(t *testing.T) {
	h := setupCategoryHandler()

	tests := []struct {
		name    string
		method  string
		path    string
		handler http.HandlerFunc
	}{
		{
			name:    "GET on CreateCategory",
			method:  http.MethodGet,
			path:    categoriesPath,
			handler: h.CreateCategory,
		},
		{
			name:    "POST on GetCategory",
			method:  http.MethodPost,
			path:    categoryByIDPath,
			handler: h.GetCategory,
		},
		{
			name:    "DELETE on ListCategories",
			method:  http.MethodDelete,
			path:    categoriesPath,
			handler: h.ListCategories,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			tt.handler(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s: status = %d, want %d", tt.name, w.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// seedCategory creates a category using the handler's CreateCategory endpoint.
// The mockCategoryIDGen always returns testCategoryID, so only one ID is used
// per handler instance. For multi-category tests, use seedCategoryOnHandler.
func seedCategory(t *testing.T, h *CategoryHandler, name string, userID string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"name": name})
	req := withCategoryUserID(httptest.NewRequest(http.MethodPost, categoriesPath, bytes.NewReader(body)), userID)
	w := httptest.NewRecorder()
	h.CreateCategory(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("seedCategory(%q, %q) failed with status %d: %s", name, userID, w.Code, w.Body.String())
	}
}

// seedCategoryViaRepo inserts a category directly into the repository, bypassing
// the handler. Useful for tests that need to pre-configure the repository state
// (e.g., marking as in-use) before calling a handler.
func seedCategoryViaRepo(repo *repository.MemoryCategoryRepository, id, name, userID string) {
	cat := &domain.Category{
		ID:        id,
		UserID:    userID,
		Name:      name,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	_ = repo.Create(context.Background(), cat)
}

// fixedIDGen returns a predetermined ID, useful for seeding multiple categories
// in the same repository with distinct IDs.
type fixedIDGen struct {
	id string
}

func (f *fixedIDGen) Generate() string {
	return f.id
}

// sequentialIDGen cycles through a list of IDs in order.
type sequentialIDGen struct {
	ids   []string
	index int
}

func (s *sequentialIDGen) Generate() string {
	id := s.ids[s.index%len(s.ids)]
	s.index++
	return id
}

// setupCategoryHandlerWithIDs creates a CategoryHandler backed by a single
// in-memory repository that will issue the provided IDs in sequence.
func setupCategoryHandlerWithIDs(ids ...string) *CategoryHandler {
	repo := repository.NewMemoryCategoryRepository()
	idGen := &sequentialIDGen{ids: ids}
	svc := service.NewCategoryService(repo, idGen)
	return NewCategoryHandler(svc)
}

// seedCategoryOnHandler seeds a category using the handler endpoint, expecting
// the handler's IDGen to produce the given expectedID. This helper is only
// useful when the caller controls the IDGen sequence (setupCategoryHandlerWithIDs).
func seedCategoryOnHandler(t *testing.T, h *CategoryHandler, name, userID, expectedID string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"name": name})
	req := withCategoryUserID(httptest.NewRequest(http.MethodPost, categoriesPath, bytes.NewReader(body)), userID)
	w := httptest.NewRecorder()
	h.CreateCategory(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("seedCategoryOnHandler(%q) failed with status %d: %s", name, w.Code, w.Body.String())
	}
}
