package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"lab1/internal/app/config"
	"lab1/internal/app/ds"
	"lab1/internal/app/role"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetSpeakers(ctx *gin.Context) {
	var speakers []ds.Speaker
	var err error
	var count int64

	searchQuery := ctx.Query("query") // получаем значение из нашего поля
	if searchQuery == "" {            // если поле поиска пусто, то просто получаем из репозитория все записи
		speakers, err = h.Repository.GetSpeakers()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		speakers, err = h.Repository.GetSpeakersByTitle(searchQuery) // в ином случае ищем заказ по заголовку
		if err != nil {
			logrus.Error(err)
		}
	}

	count = h.Repository.GetCartCount()

	fmt.Print(count)

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"speakers":   speakers,
		"cart_count": count,
	})
}

func (h *Handler) GetSpeaker(ctx *gin.Context) {
	idStr := ctx.Param("id") // получаем id заказа из урла (то есть из /speaker/:id)
	// через двоеточие мы указываем параметры, которые потом сможем считать через функцию выше
	id, err := strconv.Atoi(idStr) // так как функция выше возвращает нам строку, нужно ее преобразовать в int
	if err != nil {
		logrus.Error(err)
	}

	speaker, err := h.Repository.GetSpeaker(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "speaker.html", gin.H{
		"speaker": speaker,
	})
}

type loginReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResp struct {
	ExpiresIn   time.Duration `json:"expires_in"`
	AccessToken string        `json:"access_token"`
	TokenType   string        `json:"token_type"`
}

func (h *Handler) Login(gCtx *gin.Context) {

	cfg := &config.Config{
		JWT: config.JWTConfig{
			SigningMethod: jwt.SigningMethodHS256,
			Token:         "mysecretkey123", // Используйте это как секретный ключ
			ExpiresIn:     time.Hour * 24,   // Время жизни токена
		},
	}

	req := &loginReq{}

	// Декодируем тело запроса
	err := json.NewDecoder(gCtx.Request.Body).Decode(req)
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.GetUserByLogin(req.Login)
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	fmt.Printf("Login: %s, Password: %s\n", user.Login, user.Password)
	fmt.Printf("Login: %s, Password: %s\n", req.Login, req.Password)

	// Хешируем введённый пароль
	hashedPassword := generateHashString(req.Password)
	// Сравниваем хеши
	if hashedPassword == user.Password {
		// Определим длительность истечения срока действия токена
		var expiresDuration time.Duration
		if cfg.JWT.ExpiresIn == 0 {
			// Значение по умолчанию, например, 24 часа
			expiresDuration = time.Hour * 24
		} else {
			expiresDuration = cfg.JWT.ExpiresIn
		}

		// Генерируем JWT токен
		token := jwt.NewWithClaims(cfg.JWT.SigningMethod, &ds.JWTClaims{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(expiresDuration).Unix(),
				IssuedAt:  time.Now().Unix(),
				Issuer:    "bitop-admin",
			},
			UserUUID: uuid.New(),
			Scopes:   []string{}, // области доступа
		})

		if token == nil {
			gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("token is nil"))
			return
		}

		strToken, err := token.SignedString([]byte(cfg.JWT.Token))
		if err != nil {
			gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cant create str token"))
			return
		}

		// Отправляем ответ
		gCtx.JSON(http.StatusOK, loginResp{
			ExpiresIn:   expiresDuration, // тут тоже можно оставить исходное значение
			AccessToken: strToken,
			TokenType:   "Bearer",
		})
		log.Println(strToken)
		return
	}

	// Неверный логин или пароль
	gCtx.AbortWithStatus(http.StatusForbidden)
}

type registerReq struct {
	Login    string `json:"login"` // лучше назвать то же самое что login
	Password string `json:"password"`
}

type registerResp struct {
	Ok bool `json:"ok"`
}

func (h *Handler) Register(gCtx *gin.Context) {
	log.Println("Обработка /sign_up")
	req := &registerReq{}

	log.Printf("Incoming registration: %+v", req)

	err := json.NewDecoder(gCtx.Request.Body).Decode(req)
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if req.Password == "" {
		gCtx.AbortWithError(http.StatusBadRequest, fmt.Errorf("pass is empty"))
		return
	}

	if req.Login == "" {
		gCtx.AbortWithError(http.StatusBadRequest, fmt.Errorf("name is empty"))
		return
	}

	err = h.Repository.Register(&ds.User{
		ID:       uuid.New(),
		Role:     role.Buyer,
		Login:    req.Login,
		Password: generateHashString(req.Password), // пароли делаем в хешированном виде и далее будем сравнивать хеши, чтобы их не угнали с базой вместе
	})
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	gCtx.JSON(http.StatusOK, &registerResp{
		Ok: true,
	})
}

func generateHashString(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func (h *Handler) Logout(gCtx *gin.Context) {
	// получаем заголовок

	jwtStr := gCtx.GetHeader("Authorization")

	if !strings.HasPrefix(jwtStr, jwtPrefix) { // если нет префикса то нас дурят!
		gCtx.AbortWithStatus(http.StatusBadRequest) // отдаем что нет доступа

		return // завершаем обработку
	}

	// отрезаем префикс

	jwtStr = jwtStr[len(jwtPrefix):]

	log.Println("TEST")
	log.Println("TEST")
	log.Println("TEST")
	log.Println(jwtStr)
	token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.config.JWT.Token), nil
	})
	if err != nil {
		log.Println("Error parsing token:", err)
	} else {
		log.Println("Token parsed successfully:", token)
	}
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		log.Println(err)

		return
	}

	// сохраняем в блеклист редиса
	err = h.redis.WriteJWTToBlacklist(gCtx.Request.Context(), jwtStr, h.config.JWT.ExpiresIn)
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, err)

		log.Println("NOT END")
		return
	}

	log.Println("END")
	gCtx.Status(http.StatusOK)
}

// Ping godoc
// @Summary      Show hello text
// @Description  very very friendly response
// @Tags         Tests
// @Produce      json
// @Success      200  {object}  pingResp
// @Router       /ping/{name} [get]
func (h *Handler) Ping(gCtx *gin.Context) {
	name := gCtx.Param("name")
	gCtx.String(http.StatusOK, "Hello %s", name)
}
