package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/user"

	cnst "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
)

func Auth(cfg *config.Configuration, service user.IService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			accessTokenHeader := r.Header.Get(cnst.AccessToken)
			if accessTokenHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			aToken, err := jwt.Parse(accessTokenHeader, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(cfg.Secret.Key), nil
			})
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if claims, ok := aToken.Claims.(jwt.MapClaims); ok && aToken.Valid {
				// error if Access-Token is expired
				expired, ok2 := claims[cnst.JwtClaimsAttrTokenExpire].(float64)
				if !ok2 {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				if expired < float64(time.Now().Unix()) {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				// get user ID
				userID, ok2 := claims[cnst.JwtClaimsAttrUserID].(float64)
				if !ok2 {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				ctxWithUserID := context.WithValue(
					r.Context(),
					cnst.UserIDContextKey,
					int64(userID),
				)

				// get user Email
				ok2 = false
				userEmail, ok2 := claims[cnst.JwtClaimsAttrUserEmail]
				if !ok2 {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				ctxWithUserEmailAndUserID := context.WithValue(
					ctxWithUserID,
					cnst.UserEmailContextKey,
					userEmail,
				)

				// get user Role name
				ok2 = false
				role, ok2 := claims[cnst.JwtClaimsAttrRoleName]
				if !ok2 {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				ctxWithRoleNameAndUserEmailAndUserID := context.WithValue(
					ctxWithUserEmailAndUserID,
					cnst.RoleNameContextKey,
					role,
				)

				// get user team ID
				ok2 = false
				teamID, ok2 := claims[cnst.JwtClaimsAttrTeamID].(float64)
				if !ok2 {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				ctxWithTeamIdAndRoleNameAndUserEmailAndUserID := context.WithValue(
					ctxWithRoleNameAndUserEmailAndUserID,
					cnst.TeamIDContextKey,
					int64(teamID),
				)

				// added complex context to request
				r = r.WithContext(ctxWithTeamIdAndRoleNameAndUserEmailAndUserID)

				// check actual data for User
				user, err := service.GetUser(context.Background(), int64(userID))
				if err != nil {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				if user.Email != userEmail || user.Role.Name != role || user.Team.ID != int64(teamID) {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func AuthSuperUser() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value(cnst.RoleNameContextKey)

			if role.(string) != cnst.SuperUserRole {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func AuthAdmin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value(cnst.RoleNameContextKey)

			if role.(string) != cnst.AdminRole {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func AuthNotCaptain() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value(cnst.RoleNameContextKey)

			if role.(string) == cnst.CaptainRole {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func AuthCaptain() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value(cnst.RoleNameContextKey)

			if role.(string) != cnst.CaptainRole {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func AuthSuperUserAdminCapitan() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value(cnst.RoleNameContextKey)

			if rl, ok := role.(string); !ok || (rl != cnst.SuperUserRole && rl != cnst.AdminRole && rl != cnst.CaptainRole) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func AuthSuperUserAdmin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value(cnst.RoleNameContextKey)

			if rl, ok := role.(string); !ok || (rl != cnst.SuperUserRole && rl != cnst.AdminRole) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func AuthSuperUserAdminTournamentMaster() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value(cnst.RoleNameContextKey)

			if rl, ok := role.(string); !ok || (rl != cnst.SuperUserRole && rl != cnst.AdminRole && rl != cnst.TournamentMaster) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func AuthSuperUserAdminCapitanTournamentMaster() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value(cnst.RoleNameContextKey)

			if rl, ok := role.(string); !ok ||
				(rl != cnst.SuperUserRole && rl != cnst.AdminRole && rl != cnst.CaptainRole && rl != cnst.TournamentMaster) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func AuthTournamentMaster() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value(cnst.RoleNameContextKey)

			if rl, ok := role.(string); !ok || (rl != cnst.TournamentMaster) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
