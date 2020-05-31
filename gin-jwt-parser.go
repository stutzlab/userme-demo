package main

import (
	"fmt"
	"regexp"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/flaviostutz/go-utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//Config configuration properties for JWT Parser
type Config struct {
	//SkipPathRegex Request paths (as in gin.Context.FullPath()) that matches this regex won't be processed
	SkipPathRegex  string
	skipPathRegexp *regexp.Regexp

	//FromBearer Name of HTTP Header to load JWT token from. Header value should be prefixed by "Bearer "
	FromBearer string
	//FromCookie Name of the cookie to load JWT token from
	FromCookie string
	//FromQuery Name of request query param to load JWT token from
	FromQuery string

	//JWTSigningMethod JWT signing method. One of HS*, ES* or RS*
	JWTSigningMethod string
	//JWTVerifyKeyFile JWT signing file path (if ES or RS, must contain a public key)
	JWTVerifyKeyFile string
	//JWTContextName Name of the context property to place JWT claims after token is parsed and validated. defaults to 'jwt'
	// JWTContextName string
	jwtVerifyKey interface{}

	//RequiredIssuer Required 'iss' value in token. Not verified if empty.
	RequiredIssuer string
	//RequiredType Required 'typ' value in token. Not verified if empty
	RequiredType string
	//RequiredClaims Required values in JWT token claims. No effect if empty.
	RequiredClaims map[string]string
}

//Middleware Analyses http request, parse existing JWT tokens and set object "jwt" to gin context according to configuration.Middleware
//The jwt token claims can be later checked by request handlers with "c.GetString(...)"
func Middleware(config Config) gin.HandlerFunc {
	// if config.JWTContextName == "" {
	// 	config.JWTContextName = "jwt"
	// }

	if config.JWTSigningMethod == "" {
		config.JWTSigningMethod = "ES256"
	}

	if config.SkipPathRegex != "" {
		config.skipPathRegexp = regexp.MustCompile(config.SkipPathRegex)
	}

	if config.FromBearer == "" && config.FromCookie == "" && config.FromQuery == "" {
		panic("gin-jwt-parser: One of FromBearer, FromCookie or FromQuery config must be defined")
	}

	if !strings.HasPrefix(config.JWTSigningMethod, "HS") && !strings.HasPrefix(config.JWTSigningMethod, "RS") && !strings.HasPrefix(config.JWTSigningMethod, "ES") {
		panic("gin-jwt-parser: JWTSigningMethod must be HS*, ES* or RS*")
	}

	if config.JWTVerifyKeyFile == "" {
		panic("gin-jwt-parser: JWTVerifyKeyFile is required")
	}

	pubk, err := utils.ParseKeyFromPEM(config.JWTVerifyKeyFile, false)
	if err != nil {
		panic(fmt.Sprintf("gin-jwt-parser: Error parsing JWT pub key. err=%s", err))
	}
	config.jwtVerifyKey = pubk

	return func(ctx *gin.Context) {

		if config.skipPathRegexp != nil {
			if config.skipPathRegexp.MatchString(ctx.FullPath()) {
				logrus.Debugf("Skipping JWT token parser for path %s", ctx.FullPath)
				return
			}
		}

		jwtStr := ""
		if config.FromCookie != "" {
			v, err := ctx.Cookie(config.FromCookie)
			if err == nil {
				logrus.Debugf("Using token from Cookie %s", config.FromCookie)
				jwtStr = v
			}
		}
		if config.FromQuery != "" {
			v, exists := ctx.GetQuery(config.FromQuery)
			if exists {
				logrus.Debugf("Using token from Query param %s", config.FromQuery)
				jwtStr = v
			}
		}
		if config.FromBearer != "" {
			v := ctx.GetHeader(config.FromBearer)
			c := strings.Split(v, "Bearer ")
			if len(c) > 1 {
				if c[1] != "" {
					logrus.Debugf("Using token from HTTP Header %s", config.FromBearer)
					jwtStr = c[1]
				}
			}
		}

		if jwtStr == "" {
			ctx.AbortWithStatusJSON(401, gin.H{"message": "JWT token is required"})
			return
		}

		token, err := jwt.Parse(jwtStr, func(token *jwt.Token) (interface{}, error) {
			// if token.Method.Alg() != config.JWTSigningMethod {
			// 	return nil, fmt.Errorf("Invalid JWT signing algorithm found. found=%s required=%s", token.Method.Alg(), config.JWTSigningMethod)
			// }
			return config.jwtVerifyKey, nil
		})
		if err != nil {
			logrus.Debugf("Couldn't parse JWT token. err=%s", err)
			ctx.AbortWithStatusJSON(401, gin.H{"message": "JWT token is invalid"})
			return
		}

		if !token.Valid {
			logrus.Debugf("JWT token is invalid")
			ctx.AbortWithStatusJSON(401, gin.H{"message": "JWT token is invalid"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logrus.Warnf("Couldn't load token claims")
			ctx.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		// logrus.Debugf("JWT token loaded. Validating.")

		if config.RequiredIssuer != "" {
			iss, exists := claims["iss"]
			if !exists {
				logrus.Debugf("JWT iss claim not found")
				ctx.AbortWithStatusJSON(401, gin.H{"message": "JWT token is invalid"})
				return
			}
			if iss != config.RequiredIssuer {
				logrus.Debugf("Wrong JWT issuer. required=%s found=%s", config.RequiredIssuer, iss)
				ctx.AbortWithStatusJSON(401, gin.H{"message": "JWT token is invalid"})
				return
			}
		}

		if config.RequiredType != "" {
			typ, exists := claims["typ"]
			if !exists {
				logrus.Debugf("JWT typ claim not found")
				ctx.AbortWithStatusJSON(401, gin.H{"message": "JWT token is invalid"})
				return
			}
			if typ != config.RequiredType {
				logrus.Debugf("Wrong JWT type. required=%s found=%s", config.RequiredType, typ)
				ctx.AbortWithStatusJSON(401, gin.H{"message": "JWT token is invalid"})
				return
			}
		}

		if config.RequiredClaims != nil {
			for rk, rv := range config.RequiredClaims {
				cv, _ := claims[rk]
				if cv != rv {
					logrus.Debugf("Required claim not found. claim=%s requiredValue=%s foundValue=%s", rk, rv, cv)
					ctx.AbortWithStatusJSON(403, gin.H{"message": "JWT claim not found"})
					return
				}

			}
		}

		for k, v := range claims {
			ctx.Set(k, v)
		}

		logrus.Debugf("JWT token claims set to context")
	}
}
