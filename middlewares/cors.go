package middlewares

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware memantulkan Origin request bila diizinkan.
//
// CORS_ALLOWED_ORIGINS = daftar origin dipisah koma.
//   - kosong  -> permisif: pantulkan origin apa pun (dev). Ini valid dengan
//                credentials, tidak seperti "*" yang ditolak browser saat
//                Allow-Credentials: true.
//   - terisi  -> hanya origin yang terdaftar yang dipantulkan (produksi).
func CORSMiddleware() gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, o := range strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			allowed[o] = true
		}
	}
	allowAll := len(allowed) == 0 // kosong = izinkan semua (dev); isi env untuk kunci di prod

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && (allowAll || allowed[origin]) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, HEAD, PATCH, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
