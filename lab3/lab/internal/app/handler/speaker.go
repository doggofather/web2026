package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"lab1/internal/app/ds"
	"lab1/internal/app/role"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/minio/minio-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) GetSpeakers(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var speakers []ds.Speaker
	var err error
	var count int64

	searchQuery := c.Query("query")
	if searchQuery == "" {
		speakers, err = h.Repository.GetSpeakers()
		if err != nil {
			logrus.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		speakers, err = h.Repository.GetSpeakersByTitle(searchQuery)
		if err != nil {
			logrus.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Предполагаем, что GetCartCount адаптирован для пользователя (если нет, замените на h.Repository.GetCartCount(uint(userID.(uint))))
	count = h.Repository.GetCartCount()

	fmt.Print(count)

	c.JSON(http.StatusOK, gin.H{
		"speakers":   speakers,
		"cart_count": count,
	})
}

func (h *Handler) GetSpeaker(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	speaker, err := h.Repository.GetSpeaker(id)
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"speaker": speaker,
	})
}

// Новые функции для Speakers
func (h *Handler) CreateSpeaker(c *gin.Context) {
	var speaker ds.Speaker
	if err := c.ShouldBindJSON(&speaker); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// УБЕРИТЕ КОММЕНТАРИИ (СИМВОЛЫ //) НИЖЕ:
	if err := h.Repository.CreateSpeaker(&speaker); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, speaker)
}

func (h *Handler) UpdateSpeaker(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var speaker ds.Speaker
	if err := c.ShouldBindJSON(&speaker); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	speaker.ID = uint(id)
	if err := h.Repository.UpdateSpeaker(speaker.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, speaker)
}

func (h *Handler) DeleteSpeakerAPI(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	speaker, err := h.Repository.GetSpeaker(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Speaker not found"})
		return
	}

	// Удаление из MinIO (Старая версия)
	if speaker.Img != "" && minioClient != nil {
		// Извлекаем только имя файла из URL (например, из http://.../4.png получим 4.png)
		fileName := filepath.Base(speaker.Img)

		// want (string, string) -> bucketName, objectName
		err := minioClient.RemoveObject("speakers", fileName)
		if err != nil {
			fmt.Println("MinIO delete error:", err)
		}
	}

	if err := h.Repository.DeleteSpeakerHard(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Speaker deleted successfully"})
}

func (h *Handler) AddSpeakerToDraft(c *gin.Context) {
	userID, _ := c.Get("user_id")
	speakerID, _ := strconv.Atoi(c.Param("id"))
	if err := h.Repository.AddSpeakerToDraft(uint(userID.(uint)), uint(speakerID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Added to draft"})
}

func (h *Handler) UploadSpeakerImage(c *gin.Context) {
	// Проверяем инициализацию (поле в репозитории должно быть Minio с большой буквы)
	if h.Repository.Minio == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MinIO client is nil"})
		return
	}

	speakerID, _ := strconv.Atoi(c.Param("id"))
	imgID := c.Param("img_id")

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
		return
	}
	defer file.Close()

	newFileName := imgID + filepath.Ext(header.Filename)

	// ВАЖНО: берем бакет из репозитория (там теперь "speakers")
	_, err = h.Repository.Minio.PutObject(
		h.Repository.Minio_bucket_name,
		newFileName,
		file,
		header.Size,
		minio.PutObjectOptions{ContentType: header.Header.Get("Content-Type")},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MinIO upload failed: " + err.Error()})
		return
	}

	// Ссылка для БД
	newURL := fmt.Sprintf("http://localhost:9000/speakers/%s", newFileName)

	if err := h.Repository.UpdateSpeakerImage(uint(speakerID), newURL); err != nil {
		// Выводим реальную причину (например, "column pic does not exist")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "url": newURL})
}

// Новые функции для Users
// Define a separate input struct for registration (to accept password)
type RegisterInput struct {
	Login       string `json:"login" binding:"required"`
	Password    string `json:"password" binding:"required"`
	IsModerator bool   `json:"is_moderator"`
}

func (h *Handler) RegisterUser(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create the user with the hashed password
	user := ds.User{
		Login:    input.Login,
		Password: string(hashedPassword), // Store the hash
	}

	// Save to DB via repository
	if err := h.Repository.RegisterUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the user (password will be hidden due to JSON tag)
	c.JSON(http.StatusCreated, user)
}

func (h *Handler) DeauthenticateUser(c *gin.Context) {
	// Для JWT деавторизация обычно обрабатывается на клиенте (удаление токена)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

func (h *Handler) GetUserProfile(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		// Обработка ошибки: если id не число, вернуть 400 Bad Request
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	user, err := h.Repository.GetUserProfile(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) UpdateUserProfile(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		// Обработка ошибки: если id не число, вернуть 400 Bad Request
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	var user ds.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user.UUID = uint(userID)
	if err := h.Repository.UpdateUserProfile(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User info updated"})
}

type registerReq struct {
	Login    string `json:"login"` // лучше назвать то же самое что login
	Password string `json:"password"`
}

type registerResp struct {
	Ok bool `json:"ok"`
}

func generateHashString(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
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
		Role:     role.Guest,
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
	req := &loginReq{}

	// 1. Декодируем тело запроса
	if err := json.NewDecoder(gCtx.Request.Body).Decode(req); err != nil {
		gCtx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// 2. Ищем пользователя в базе
	user, err := h.Repository.GetUserByLogin(req.Login)
	if err != nil {
		// Если пользователь не найден или ошибка БД
		gCtx.JSON(http.StatusForbidden, gin.H{"error": "invalid login or password"})
		return
	}

	// 3. Проверяем пароль
	hashedPassword := generateHashString(req.Password)

	logrus.Infof("DB User: %s, DB Pass Hash: %s", user.Login, user.Password)
	logrus.Infof("Input Login: %s, Input Pass Hash: %s", req.Login, hashedPassword)

	if hashedPassword != user.Password {
		gCtx.JSON(http.StatusForbidden, gin.H{"error": "password mismatch"}) // Временно измени текст, чтобы понять, что это именно пароль
		return
	}

	// 4. Настраиваем время жизни токена
	expiresDuration := h.config.JWT.ExpiresIn
	if expiresDuration == 0 {
		expiresDuration = time.Hour * 24 // Значение по умолчанию
	}

	// 5. Генерируем Claims
	// ВНИМАНИЕ: Используем user.UUID из базы данных, а не uuid.New()!
	claims := &ds.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiresDuration).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "bitop-admin",
		},
		UserUUID: user.UUID, // Поле должно называться так же, как в структуре JWTClaims
		Scopes:   []string{},
	}

	// 6. Создаем и подписываем токен ключом из h.config
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	key := h.config.JWT.Token
	if key == "" {
		key = "mysecretkey123"
	}
	strToken, err := token.SignedString([]byte(key))
	// Используем общий секретный ключ h.config.JWT.Token
	if err != nil {
		logrus.Errorf("failed to sign token: %v", err)
		gCtx.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	// 7. Отправляем успешный ответ
	gCtx.JSON(http.StatusOK, loginResp{
		ExpiresIn:   expiresDuration,
		AccessToken: strToken,
		TokenType:   "Bearer",
	})
}

func isImage(contentType string) bool {
	imageTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, t := range imageTypes {
		if contentType == t {
			return true
		}
	}
	return false
}

// Вспомогательная функция, выполняющая валидацию загруженного файла
func validateFileUpload(header *multipart.FileHeader) (int, error) {
	// Окрываем чтение файлового потока
	file, err := header.Open()

	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("не удалось получить файл")
	}

	defer file.Close()

	// Знакомая логика определения типа содержимого...

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)

	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("не удалось прочитать файл")
	}

	contentType := http.DetectContentType(buffer)

	if !isImage(contentType) {
		return http.StatusBadRequest, fmt.Errorf("файл должен быть изображением")
	}

	// Позиция чтения файла возвращается к исходной
	_, err = file.Seek(0, 0)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("ошибка при обработке файла")
	}

	return 0, nil
}
func (h *Handler) AddSpeakerToMeetup(c *gin.Context) {
	// Достаем ID из параметров пути
	mID, errM := strconv.Atoi(c.Param("meetup_id"))
	sID, errS := strconv.Atoi(c.Param("speaker_id"))

	if errM != nil || errS != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid IDs in URL"})
		return
	}

	// Создаем структуру для базы
	binding := ds.SpeakerMeetup{
		MeetupID:  uint(mID),
		SpeakerID: uint(sID),
	}

	// Сохраняем в БД
	if err := h.Repository.CreateSpeakerMeetup(&binding); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, binding)
}

func (h *Handler) Logout(gCtx *gin.Context) {
	jwtStr := strings.TrimPrefix(gCtx.GetHeader("Authorization"), "Bearer ")

	// Вместо ParseWithClaims (который проверяет подпись), используем ParseUnverified
	// Нам нужно просто достать Claims, чтобы понять, живой ли он еще (опционально)
	claims := &ds.JWTClaims{}
	_, _, err := new(jwt.Parser).ParseUnverified(jwtStr, claims)

	if err != nil {
		gCtx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid token format"})
		return
	}

	// Теперь просто пишем его в Redis.
	// Даже если подпись "битая", мы его баним, и он больше не пройдет через Middleware
	err = h.redis.WriteJWTToBlacklist(gCtx.Request.Context(), jwtStr, h.config.JWT.ExpiresIn)
	if err != nil {
		gCtx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "redis error"})
		return
	}

	gCtx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
