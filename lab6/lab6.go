package main

import (
	"net/http"
    "strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Book struct {
	// TODO: Finish struct
	ID    int    `json:"id"`
    Name  string `json:"name"`
    Pages int    `json:"pages"`
}

var bookshelf = []Book{
	// TODO: Init bookshelf
	{ID: 1, Name: "Blue Bird", Pages: 500},
}

var curID = 1

func getBooks(c *gin.Context) {
	c.JSON(http.StatusOK, bookshelf)
}

func getBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "book not found"})
        return
    }

    for _, book := range bookshelf {
        if book.ID == id {
            c.JSON(http.StatusOK, book)
            return
        }
    }

    c.JSON(http.StatusNotFound, gin.H{"message": "book not found"})
}

func addBook(c *gin.Context) {
	var newBook struct {
        Name  string `json:"name"`
        Pages int    `json:"pages"`
    }
    if err := c.BindJSON(&newBook); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
        return
    }

	// 檢查書名是否重複
    for _, book := range bookshelf {
        if strings.ToLower(book.Name) == strings.ToLower(newBook.Name) {
            c.JSON(http.StatusConflict, gin.H{"message": "duplicate book name"})
            return
        }
    }
    
    curID++
    bookToAdd := Book{
        ID:    curID,
        Name:  newBook.Name,
        Pages: newBook.Pages,
    }

    bookshelf = append(bookshelf, bookToAdd)
    c.JSON(http.StatusCreated, bookToAdd)
}

func deleteBook(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    // println(id)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid book ID"})
        return
    }

    for index, book := range bookshelf {
        // println(book.ID, "   ", id)
        if book.ID == id {
            bookshelf = append(bookshelf[:index], bookshelf[index+1:]...)
            c.Status(http.StatusNoContent)
            return
        }
    }

    // 返回 204，即使 ID 不存在
    c.Status(http.StatusNoContent)
}


func updateBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid book ID"})
        return
    }

    var updatedBook Book
    if err := c.BindJSON(&updatedBook); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
        return
    }

    for i, book := range bookshelf {
        if book.ID == id {
            // 更新書籍時檢查重複書名
            for _, b := range bookshelf {
                if b.ID != id && strings.ToLower(b.Name) == strings.ToLower(updatedBook.Name) {
                    c.JSON(http.StatusConflict, gin.H{"message": "duplicate book name"})
                    return
                }
            }

            bookshelf[i] = updatedBook
            bookshelf[i].ID = id // 保持 ID 不變
            c.JSON(http.StatusOK, bookshelf[i])
            return
        }
    }

    c.JSON(http.StatusNotFound, gin.H{"message": "book not found"})
}

func main() {
	r := gin.Default()
	r.RedirectFixedPath = true

	// TODO: Add routes
	r.GET("/bookshelf", getBooks)
    r.GET("/bookshelf/:id", getBook)
    r.POST("/bookshelf", addBook)
    r.DELETE("/bookshelf/:id", deleteBook)
    r.PUT("/bookshelf/:id", updateBook)

	err := r.Run(":8087")
	if err != nil {
		return
	}
}