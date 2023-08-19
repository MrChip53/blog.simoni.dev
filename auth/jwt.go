package auth

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"net/http"
	"os"
)

type JwtPayload struct {
	Username string `json:"username"`
	Admin    bool   `json:"admin"`
	UserId   uint   `json:"userId"`
}

type JwtRefreshPayload struct {
	JwtToken string `json:"jwtToken"`
}

func ExtractAuth(ctx *gin.Context) (jwtPayload *JwtPayload, err error) {
	jwtCookie, err := ctx.Request.Cookie("token")
	if err != nil || jwtCookie == nil {
		return refreshTokens(ctx)
	}

	return VerifyJwtToken(jwtCookie.Value)
}

func AddAuthCookies(ctx *gin.Context, jwtToken string, refreshToken string) {
	jwtCookie := http.Cookie{
		Name:     "token",
		Value:    jwtToken,
		Path:     "/",
		Domain:   "",
		MaxAge:   60,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
		HttpOnly: true,
	}
	http.SetCookie(ctx.Writer, &jwtCookie)

	refreshCookie := http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Path:     "/",
		Domain:   "",
		MaxAge:   60 * 60 * 3,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
		HttpOnly: true,
	}
	http.SetCookie(ctx.Writer, &refreshCookie)
}

func VerifyRefreshToken(token string) (payload *JwtRefreshPayload, err error) {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	return verifyRefreshToken(token, jwtSecret)
}

func VerifyJwtToken(token string) (payload *JwtPayload, err error) {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	return verifyJwtToken(token, jwtSecret)
}

func verifyJwtToken(token string, jwtSecret []byte) (payload *JwtPayload, err error) {
	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		fmt.Println("Token parsing error:", err)
		return nil, err
	}

	if _, ok := jwtToken.Claims.(jwt.MapClaims); ok && jwtToken.Valid {
		payload = &JwtPayload{
			Username: jwtToken.Claims.(jwt.MapClaims)["username"].(string),
			Admin:    jwtToken.Claims.(jwt.MapClaims)["admin"].(bool),
			UserId:   uint(jwtToken.Claims.(jwt.MapClaims)["userId"].(float64)),
		}
		return payload, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func verifyRefreshToken(token string, jwtSecret []byte) (payload *JwtRefreshPayload, err error) {
	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		fmt.Println("Token parsing error:", err)
		return nil, err
	}

	if _, ok := jwtToken.Claims.(jwt.MapClaims); ok && jwtToken.Valid {
		payload = &JwtRefreshPayload{
			JwtToken: jwtToken.Claims.(jwt.MapClaims)["jwtToken"].(string),
		}
		return payload, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func GenerateTokens(payload *JwtPayload) (token string, refreshToken string, err error) {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	token, err = generateJwtToken(payload, jwtSecret)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = generateRefreshToken(JwtRefreshPayload{JwtToken: token}, jwtSecret)
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}

func generateRefreshToken(payload JwtRefreshPayload, jwtSecret []byte) (string, error) {
	claims := jwt.MapClaims{}
	claims["jwtToken"] = payload.JwtToken

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtSecret)
	return tokenString, err
}

func generateJwtToken(payload *JwtPayload, jwtSecret []byte) (string, error) {
	claims := jwt.MapClaims{}
	claims["username"] = payload.Username
	claims["admin"] = payload.Admin
	claims["userId"] = payload.UserId

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtSecret)
	return tokenString, err
}

func refreshTokens(ctx *gin.Context) (jwtPayload *JwtPayload, err error) {
	refreshCookie, err := ctx.Request.Cookie("refreshToken")
	if err != nil {
		return nil, err
	}

	refreshPayload, err := VerifyRefreshToken(refreshCookie.Value)
	if err != nil {
		return nil, err
	}

	jwtPayload, err = VerifyJwtToken(refreshPayload.JwtToken)
	if err != nil {
		return nil, err
	}

	jwtToken, refreshToken, err := GenerateTokens(jwtPayload)
	if err != nil {
		return nil, err
	}

	AddAuthCookies(ctx, jwtToken, refreshToken)

	return jwtPayload, nil
}
