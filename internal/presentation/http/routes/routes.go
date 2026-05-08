// internal/presentation/http/routes/routes.go
package routes

import (
	"net/http"

	"github.com/farmanexo/auth-service/internal/presentation/http/controllers"
	"github.com/farmanexo/auth-service/internal/presentation/http/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

// SetupRoutes configura todas las rutas del servicio
func SetupRoutes(
	authController *controllers.AuthController,
	legalController *controllers.LegalController,
	authMiddleware *middlewares.AuthMiddleware,
	rateLimitMiddleware *middlewares.RateLimitMiddleware,
) *chi.Mux {
	r := chi.NewRouter()

	// ========================================
	// MIDDLEWARES GLOBALES
	// ========================================

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://farmanexo.pe"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middlewares.CorrelationID)

	// ========================================
	// SWAGGER DOCUMENTATION
	// ========================================

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:4001/swagger/doc.json"),
	))

	// ========================================
	// HEALTH CHECK
	// ========================================

	r.Get("/health", authController.HealthCheck)
	r.Get("/auth/health", authController.HealthCheck)
	r.Get("/", authController.HealthCheck)

	// ========================================
	// API ROUTES - VERSION 1
	// ========================================
	//
	// Principio (services/CLAUDE.md §5.5): los recursos viven bajo prefijos
	// plurales por dominio. /legal NO es un dominio top-level — es un
	// sub-recurso de /auth (los términos legales son cosa de la cuenta y
	// del consentimiento, gestionado por auth-service). Por eso vive en
	// /api/v1/auth/legal/*, no en /api/v1/legal/*.
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			// LEGAL DOCUMENTS — público (sin JWT).
			// Términos, Privacidad, etc. accesibles incluso sin login porque
			// la página /terminos debe verse antes de registrarse.
			r.Route("/legal", func(r chi.Router) {
				r.Get("/types", legalController.ListTypes)
				r.Get("/{type_code}", legalController.GetCurrent)
				r.Get("/{type_code}/versions", legalController.ListVersions)
				r.Get("/{type_code}/versions/{version}", legalController.GetByVersion)
			})

			// Rutas públicas
			r.Post("/register", authController.Register)
			r.Post("/login", authController.Login)
			r.With(rateLimitMiddleware.RefreshRateLimit).Post("/refresh", authController.RefreshToken)

			// Rutas protegidas (requieren autenticación)
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequireAuth)

				r.Post("/logout", authController.Logout)

				// Derechos ARCO (LPDP Ley 29733)
				r.Get("/me/consents", authController.GetMyConsents)
				r.Get("/me/consents/status", authController.GetConsentsStatus)
				r.Post("/me/consents/accept", authController.AcceptConsents)
				r.Delete("/me", authController.DeleteMyAccount)
			})
		})
	})

	// ========================================
	// API ROUTES - VERSION 2 (Futuro)
	// ========================================

	r.Route("/api/v2", func(r chi.Router) {
		r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("API v2 - Próximamente"))
		})
	})

	return r
}
