package service
 
import (
    "crypto/sha256"
	"encoding/hex"
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
	database "todolist.go/db"
)
 
func NewUserForm(ctx *gin.Context) {
    ctx.HTML(http.StatusOK, "new_user_form.html", gin.H{"Title": "Register user"})
}

func hash(pw string) []byte {
    const salt = "todolist.go#"
    h := sha256.New()
    h.Write([]byte(salt))
    h.Write([]byte(pw))
    return h.Sum(nil)
}

func passcheck(pw string) bool {
	cnt := 0
	for _, r := range pw {
		if '0' <= r && r <= '9' {
			cnt++
		}
	}
	if cnt == len(pw) {
		return false
	}
	return len(pw) >= 6
}

func RegisterUser(ctx *gin.Context) {
    // フォームデータの受け取り
    username := ctx.PostForm("username")
    password := ctx.PostForm("password")
	re_password := ctx.PostForm("re_password")
    switch {
    case username == "":
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Usernane is not provided", "Username": username})
    case password == "":
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is not provided", "Password": password})
	case re_password != password:
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "You entered a different password", "Username": username})
		return
	case !passcheck(password):
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password must be at least 6 characters and cannnot contain only numbers", "Username": username})
		return
    }
    
    // DB 接続
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
 
    // 重複チェック
    var duplicate int
    err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", username)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    if duplicate > 0 {
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Username is already taken", "Username": username, "Password": password, "Re_password": re_password})
        return
    }
 
    // DB への保存
    result, err := db.Exec("INSERT INTO users(name, password) VALUES (?, ?)", username, hash(password))
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
 
    // 保存状態の確認
    id, _ := result.LastInsertId()
    var user database.User
    err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", id)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    //ctx.JSON(http.StatusOK, user)
	ctx.HTML(http.StatusOK, "new_user_fin.html", gin.H{"Title": "Resister fin"})
}

func LoginForm(ctx *gin.Context) {
    ctx.HTML(http.StatusOK, "login.html", gin.H{"Title": "Login"})
}

const userkey = "user"
 
func Login(ctx *gin.Context) {
    username := ctx.PostForm("username")
    password := ctx.PostForm("password")
 
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
 
    // ユーザの取得
    var user database.User
    err = db.Get(&user, "SELECT id, name, password, is_delete FROM users WHERE name = ? AND is_delete = false", username)
    if err != nil {
        ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "No such user"})
        return
    }
 
    // パスワードの照合
    if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
        ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "Incorrect password"})
        return
    }
 
    // セッションの保存
    session := sessions.Default(ctx)
    session.Set(userkey, user.ID)
    session.Save()
 
    ctx.Redirect(http.StatusFound, "/list")
}

func LoginCheck(ctx *gin.Context) {
    if sessions.Default(ctx).Get(userkey) == nil {
        ctx.Redirect(http.StatusFound, "/login")
        ctx.Abort()
    } else {
        ctx.Next()
    }
}

func LoginCheckTaskID(ctx *gin.Context) {
    if sessions.Default(ctx).Get(userkey) == nil {
        ctx.Redirect(http.StatusFound, "/login")
        ctx.Abort()
    }
    // parse ID given as a parameter
    taskID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
    // Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
    userID := sessions.Default(ctx).Get("user")

    var owner []database.Ownership
    err = db.Select(&owner, "SELECT * FROM ownership WHERE user_id = ? AND task_id = ?", userID,  taskID)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    if len(owner) > 0 {
        ctx.Next()
    } else {
        ctx.Redirect(http.StatusFound, "/")
        ctx.Abort()
    }
}

func Logout(ctx *gin.Context) {
    session := sessions.Default(ctx)
    session.Clear()
    session.Options(sessions.Options{MaxAge: -1})
    session.Save()
    ctx.Redirect(http.StatusFound, "/")
}

func ResureDelete(ctx *gin.Context) {
    ctx.HTML(http.StatusOK, "resure_delete.html", gin.H{"Title": "Resure"})
}

func DeleteUser(ctx *gin.Context) {
    username := ctx.PostForm("username")
    password := ctx.PostForm("password")

    // Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
    
    // ユーザの取得
    var user database.User
    err = db.Get(&user, "SELECT id, name, password FROM users WHERE name = ? AND is_delete = false", username)
    if err != nil {
        ctx.HTML(http.StatusBadRequest, "resure_delete.html", gin.H{"Title": "Login", "Username": username, "Error": "No such user"})
        return
    }

    // パスワードの照合
    if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
        ctx.HTML(http.StatusBadRequest, "resure_delete.html", gin.H{"Title": "Login", "Username": username, "Error": "Incorrect password"})
        return
    }

   _, err = db.Exec("UPDATE users set is_delete = true WHERE name = ?", username)
   if err != nil {
       Error(http.StatusInternalServerError, err.Error())(ctx)
       return
   }
   ctx.Redirect(http.StatusFound, "/")
}

func EditUserForm(ctx *gin.Context) {
    ctx.HTML(http.StatusOK, "form_edit_user.html", gin.H{"Title": "edit_user"})
}

func UpdateUser(ctx *gin.Context) {
    before_username := ctx.PostForm("before_username")
    before_password := ctx.PostForm("before_password")

    // Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
    
    // ユーザの取得
    var user database.User
    err = db.Get(&user, "SELECT id, name, password FROM users WHERE name = ? AND is_delete = false", before_username)
    if err != nil {
        ctx.HTML(http.StatusBadRequest, "form_edit_user.html", gin.H{"Title": "edit_user", "BeforeUsername": before_username, "Error": "No such user"})
        return
    }

    // パスワードの照合
    if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(before_password)) {
        ctx.HTML(http.StatusBadRequest, "form_edit_user.html", gin.H{"Title": "edit_user", "BeforeUsername": before_username, "Error": "Incorrect password"})
        return
    }

    username := ctx.PostForm("after_username")
    password := ctx.PostForm("after_password")
    re_password := ctx.PostForm("re_password")
    switch {
    case username == "":
        ctx.HTML(http.StatusBadRequest, "form_edit_user.html", gin.H{"Title": "edit_user", "Error": "Usernane is not provided", "AfterUsername": username})
    case password == "":
        ctx.HTML(http.StatusBadRequest, "form_edit_user.html", gin.H{"Title": "edit_user", "Error": "Password is not provided"})
	case re_password != password:
		ctx.HTML(http.StatusBadRequest, "form_edit_user.html", gin.H{"Title": "edit_user", "Error": "You entered a different password", "BeforeUsername": before_username, "BeforePassword": before_password, "AfterUsername": username})
		return
	case !passcheck(password):
		ctx.HTML(http.StatusBadRequest, "form_edit_user.html", gin.H{"Title": "edit_user", "Error": "Password must be at least 6 characters and cannnot contain only numbers", "BeforeUsername": before_username, "BeforePassword": before_password, "AfterUsername": username})
		return
    }

    // 重複チェック
    var duplicate int
    err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", username)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    if duplicate > 0 {
        ctx.HTML(http.StatusBadRequest, "form_edit_user.html", gin.H{"Title": "edit_user", "Error": "Username is already taken", "BeforeUsername": before_username, "BeforePassword": before_password, "AfterUsername": username, "Password": password, "Re_password": re_password})
        return
    }

   _, err = db.Exec("UPDATE users set name = ?, password = ? WHERE name = ?", username, hash(password), before_username)
   if err != nil {
       Error(http.StatusInternalServerError, err.Error())(ctx)
       return
   }
   ctx.Redirect(http.StatusFound, "/")
}