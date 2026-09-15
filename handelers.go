package auth

import (
    "net/http"
    "strings"
    "sync"

    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"

    "ticket-system/internal/middleware"
)

type User struct {
    ID       int    `json:"id"`
    Username string `json:"username"`
    Password string `json:"-"`
}

type credentials struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

var userStore = struct {
    sync.RWMutex
    users  map[string]User
    nextID int
}{users: make(map[string]User), nextID: 1}

func Register(c *gin.Context) {
    var req credentials
    if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Username) == "" || req.Password == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
        return
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
        return
    }

    username := strings.TrimSpace(req.Username)
    userStore.Lock()
    defer userStore.Unlock()
    if _, exists := userStore.users[username]; exists {
        c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
        return
    }
    user := User{ID: userStore.nextID, Username: username, Password: string(hash)}
    userStore.nextID++
    userStore.users[username] = user
    c.JSON(http.StatusCreated, gin.H{"id": user.ID, "username": user.Username})
}

func Login(c *gin.Context) {
    var req credentials
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
        return
    }

    userStore.RLock()
    user, exists := userStore.users[strings.TrimSpace(req.Username)]
    userStore.RUnlock()
    if !exists || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
        return
    }

    token, err := middleware.GenerateJWT(user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create token"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"token": token})
}
