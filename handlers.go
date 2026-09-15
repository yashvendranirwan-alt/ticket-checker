package tickets

import (
    "net/http"
    "strconv"
    "strings"
    "sync"

    "github.com/gin-gonic/gin"
)

const (
    Open       = "open"
    InProgress = "in_progress"
    Closed     = "closed"
)

type Ticket struct {
    ID     int    `json:"id"`
    UserID int    `json:"user_id"`
    Title  string `json:"title"`
    Status string `json:"status"`
}

type createRequest struct {
    Title string `json:"title"`
}

type statusRequest struct {
    Status string `json:"status"`
}

var ticketStore = struct {
    sync.RWMutex
    tickets map[int]Ticket
    nextID  int
}{tickets: make(map[int]Ticket), nextID: 1}

func CreateTicket(c *gin.Context) {
    var req createRequest
    if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Title) == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
        return
    }

    ticketStore.Lock()
    ticket := Ticket{ID: ticketStore.nextID, UserID: c.GetInt("user_id"), Title: strings.TrimSpace(req.Title), Status: Open}
    ticketStore.nextID++
    ticketStore.tickets[ticket.ID] = ticket
    ticketStore.Unlock()
    c.JSON(http.StatusCreated, ticket)
}

func ListTickets(c *gin.Context) {
    userID := c.GetInt("user_id")
    ticketStore.RLock()
    tickets := make([]Ticket, 0)
    for _, ticket := range ticketStore.tickets {
        if ticket.UserID == userID {
            tickets = append(tickets, ticket)
        }
    }
    ticketStore.RUnlock()
    c.JSON(http.StatusOK, tickets)
}

func GetTicket(c *gin.Context) {
    ticket, status := findOwnedTicket(c)
    if status != http.StatusOK {
        c.JSON(status, gin.H{"error": "ticket not found or not owned by user"})
        return
    }
    c.JSON(http.StatusOK, ticket)
}

func UpdateStatus(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil || id < 1 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
        return
    }
    var req statusRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
        return
    }

    ticketStore.Lock()
    ticket, exists := ticketStore.tickets[id]
    if !exists {
        ticketStore.Unlock()
        c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
        return
    }
    if ticket.UserID != c.GetInt("user_id") {
        ticketStore.Unlock()
        c.JSON(http.StatusForbidden, gin.H{"error": "ticket belongs to another user"})
        return
    }
    if !validTransition(ticket.Status, req.Status) {
        ticketStore.Unlock()
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status transition"})
        return
    }
    ticket.Status = req.Status
    ticketStore.tickets[id] = ticket
    ticketStore.Unlock()
    c.JSON(http.StatusOK, ticket)
}

func findOwnedTicket(c *gin.Context) (Ticket, int) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil || id < 1 {
        return Ticket{}, http.StatusBadRequest
    }
    ticketStore.RLock()
    ticket, exists := ticketStore.tickets[id]
    ticketStore.RUnlock()
    if !exists {
        return Ticket{}, http.StatusNotFound
    }
    if ticket.UserID != c.GetInt("user_id") {
        return Ticket{}, http.StatusForbidden
    }
    return ticket, http.StatusOK
}

func validTransition(current, next string) bool {
    return (current == Open && next == InProgress) || (current == InProgress && next == Closed)
}