package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nillga/jwt-server/entity"
	"github.com/nillga/mehm-services-api-gateway/dto"
	"github.com/nillga/mehm-services-api-gateway/service"
	"github.com/nillga/mehm-services-api-gateway/utils"
)

type ReadController interface {
	GetAllMehms(w http.ResponseWriter, r *http.Request)
	GetSpecificMehm(w http.ResponseWriter, r *http.Request)
	GetComment(w http.ResponseWriter, r *http.Request)
}

type UserController interface {
	ResolveProfile(w http.ResponseWriter, r *http.Request)
	LikeMehm(w http.ResponseWriter, r *http.Request)
	PostComment(w http.ResponseWriter, r *http.Request)
	EditComment(w http.ResponseWriter, r *http.Request)
	EditMehm(w http.ResponseWriter, r *http.Request)
}

type PrivilegedController interface {
	DeleteUser(w http.ResponseWriter, r *http.Request)
	DeleteMehm(w http.ResponseWriter, r *http.Request)
}

type ApiGatewayController interface {
	ReadController
	UserController
	PrivilegedController
}

type controller struct {
}

var apiGatewayService = service.NewApiGatewayService()
var users = os.Getenv("USERS")
var mehms = os.Getenv("MEHMS")

func NewApiGatewayController() ApiGatewayController {
	return &controller{}
}

// GetMehms godoc
// @Summary      Returns a page of mehms
// @Description  Pagination can be handled via query params
// @Tags         mehms
// @Accept       json
// @Produce      json
// @Param        skip   query      int  false  "How many mehms will be skipped"
// @Param        take   query      int  false  "How many mehms will be taken"
// @Success      200  {object}  map[string]dto.MehmDTO{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      401  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Router       /mehms [get]
func (c *controller) GetAllMehms(w http.ResponseWriter, r *http.Request) {
	_, err := apiGatewayService.Authenticate(r.Header.Get("Authorization"))
	if err != nil {
		utils.Unauthorized(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	pr, err := http.NewRequest(r.Method, mehms+"/mehms?"+r.URL.Query().Encode(), r.Body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		utils.InternalServerError(w, err)
	}
}

// GetSpecificMehm godoc
// @Summary      Returns a specified mehm
// @Description  optionally showing info for privileged user
// @Tags         mehms
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "The ID of the requested mehm"
// @Success      200  {object}  dto.MehmDTO{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      401  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Failure      502  {object}  errors.ProceduralError
// @Router       /mehms/{id} [get]
func (c *controller) GetSpecificMehm(w http.ResponseWriter, r *http.Request) {
	user, err := apiGatewayService.Authenticate(r.Header.Get("Authorization"))
	if err != nil {
		utils.Unauthorized(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		utils.BadRequest(w, fmt.Errorf("mehm specification went wrong"))
		return
	}

	pr, err := http.NewRequest(r.Method, mehms+"/mehms/get/"+id+"?userId="+user.Id, r.Body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		utils.InternalServerError(w, err)
	}
}

// GetComment godoc
// @Summary      Used to show a specified comment
// @Description  optionally showing info for privileged user
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "The ID of the requested mehm"
// @Success      200  {object}  dto.CommentDTO{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      401  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Failure      502  {object}  errors.ProceduralError
// @Router       /comments/get/{id} [get]
func (c *controller) GetComment(w http.ResponseWriter, r *http.Request) {
	_, err := apiGatewayService.Authenticate(r.Header.Get("Authorization"))
	if err != nil {
		utils.Unauthorized(w, err)
		return
	}

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		utils.BadRequest(w, fmt.Errorf("comment specification went wrong"))
		return
	}

	pr, err := http.NewRequest("GET", mehms+"/comments/get/"+id, r.Body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		utils.InternalServerError(w, err)
	}
}

// ResolveProfile godoc
// @Summary      Receive Info about ones self
// @Description  These informations contain username, email address and id
// @Tags         user
// @Accept       json
// @Produce      json
// @Success      200  {object}  entity.User{}
// @Failure      401  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Router       /user [get]
func (c *controller) ResolveProfile(w http.ResponseWriter, r *http.Request) {
	user, err := apiGatewayService.Authenticate(r.Header.Get("Authorization"))
	if err != nil {
		utils.Unauthorized(w, err)
		return
	}

	if err = json.NewEncoder(w).Encode(user); err != nil {
		utils.InternalServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

// LikeMehm godoc
// @Summary      Used to like a specified mehm
// @Description  optionally showing info for privileged user
// @Tags         mehms
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "The ID of the requested mehm"
// @Success      200  {object}  interface{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      401  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Failure      502  {object}  errors.ProceduralError
// @Router       /mehms/{id}/like [post]
func (c *controller) LikeMehm(w http.ResponseWriter, r *http.Request) {
	user, err := apiGatewayService.Authenticate(r.Header.Get("Authorization"))
	if err != nil {
		utils.Unauthorized(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		utils.BadRequest(w, fmt.Errorf("mehm specification went wrong"))
		return
	}

	pr, err := http.NewRequest(r.Method, mehms+"/mehms/"+id+"/like?userId="+user.Id, r.Body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		utils.InternalServerError(w, err)
	}
}

// PostComment godoc
// @Summary      Used to add a new comment
// @Description  optionally showing info for privileged user
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        comment   query      string  true  "The comment"
// @Param        mehmId   query      int  true  "The mehm"
// @Success      200  {object}  interface{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      401  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Failure      502  {object}  errors.ProceduralError
// @Router       /comments/get/{id} [post]
func (c *controller) PostComment(w http.ResponseWriter, r *http.Request) {
	user, err := apiGatewayService.Authenticate(r.Header.Get("Authorization"))
	if err != nil {
		utils.Unauthorized(w, err)
		return
	}

	params := r.URL.Query()
	if !params.Has("comment") || !params.Has("mehmId") {
		utils.BadRequest(w, fmt.Errorf("request query parameters must be comment AND mehmId"))
		return
	}

	pr, err := http.NewRequest(r.Method, mehms+"/comments/new?"+r.URL.Query().Encode()+"&userId="+user.Id, r.Body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		utils.InternalServerError(w, err)
	}
}

// EditComment godoc
// @Summary      Used to edit an existing comment
// @Description  optionally showing info for privileged user
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        input   body      dto.CommentInput  true  "Input data"
// @Success      200  {object}  interface{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      401  {object}  errors.ProceduralError
// @Failure      403  {object}  errors.ProceduralError
// @Failure      422  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Failure      502  {object}  errors.ProceduralError
// @Router       /comments/update [post]
func (c *controller) EditComment(w http.ResponseWriter, r *http.Request) {
	user, err := apiGatewayService.Authenticate(r.Header.Get("Authorization"))
	if err != nil {
		utils.Unauthorized(w, err)
		return
	}

	if !user.Admin {
		utils.Forbidden(w, fmt.Errorf("not authorized"))
		return
	}

	var input dto.CommentInput

	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.UnprocessableEntity(w, fmt.Errorf("format problems"))
		return
	}

	if input.UserId, err = strconv.ParseInt(user.Id, 10, 64); err != nil {
		utils.InternalServerError(w, fmt.Errorf("could not resolve user"))
		return
	}
	input.Admin = user.Admin

	body := bytes.NewBuffer([]byte{})

	if err = json.NewEncoder(body).Encode(input); err != nil {
		utils.InternalServerError(w, fmt.Errorf("failed repeating request"))
		return
	}

	pr, err := http.NewRequest(r.Method, mehms+"/comments/update", body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		utils.InternalServerError(w, err)
	}
}

// EditMehm godoc
// @Summary      Used to delete a specified mehm
// @Description  optionally showing info for privileged user
// @Tags         mehms
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "The ID of the requested mehm"
// @Param		 description body string true "The new mehm description"
// @Param        title body string true "The new mehm title"
// @Success      200  {object}  interface{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      401  {object}  errors.ProceduralError
// @Failure      403  {object}  errors.ProceduralError
// @Failure      422  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Failure      502  {object}  errors.ProceduralError
// @Router       /mehms/{id}/update [post]
func (c *controller) EditMehm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		utils.BadRequest(w, fmt.Errorf("mehm specification went wrong"))
		return
	}

	user, err := apiGatewayService.Authenticate(r.Header.Get("Authorization"))
	if err != nil {
		utils.Unauthorized(w, err)
		return
	}

	if !user.Admin {
		utils.Forbidden(w, fmt.Errorf("not authorized"))
		return
	}

	var input dto.MehmInput

	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.UnprocessableEntity(w, fmt.Errorf("format problems"))
		return
	}

	if input.UserId, err = strconv.ParseInt(user.Id, 10, 64); err != nil {
		utils.InternalServerError(w, fmt.Errorf("could not resolve user"))
		return
	}
	input.Admin = user.Admin

	body := bytes.NewBuffer([]byte{})

	if err = json.NewEncoder(body).Encode(input); err != nil {
		utils.InternalServerError(w, fmt.Errorf("failed repeating request"))
		return
	}

	pr, err := http.NewRequest(r.Method, mehms+"/"+id+"/update", body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		utils.InternalServerError(w, err)
	}
}

// Delete godoc
// @Summary      Deletes a targeted User
// @Description  Self-delete; admins can delete anybody
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        input   body      entity.DeleteUserInput  true  "Input data"
// @Success      200  {object}  interface{}
// @Failure      401  {object}  errors.ProceduralError
// @Failure      403  {object}  errors.ProceduralError
// @Failure      405  {object}  errors.ProceduralError
// @Failure      422  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Failure      502  {object}  errors.ProceduralError
// @Router       /user/delete [post]
func (c *controller) DeleteUser(w http.ResponseWriter, r *http.Request) {
	user, err := apiGatewayService.Authenticate(r.Header.Get("Authorization"))
	if err != nil {
		utils.Unauthorized(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var deleteId entity.DeleteUserInput
	if err = json.NewDecoder(r.Body).Decode(&deleteId); err != nil {
		utils.UnprocessableEntity(w, err)
		return
	}

	if !user.Admin && user.Id != deleteId.Id {
		utils.Forbidden(w, err)
		return
	}

	pr, err := http.NewRequest(r.Method, users+"/delete?id="+deleteId.Id, r.Body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
	}
}

// RemoveMehm godoc
// @Summary      Used to delete a specified mehm
// @Description  optionally showing info for privileged user
// @Tags         mehms
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "The ID of the requested mehm"
// @Success      200  {object}  interface{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      401  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Failure      502  {object}  errors.ProceduralError
// @Router       /mehms/{id}/remove [post]
func (c *controller) DeleteMehm(w http.ResponseWriter, r *http.Request) {
	user, err := apiGatewayService.Authenticate(r.Header.Get("Authorization"))
	if err != nil {
		utils.Unauthorized(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		utils.BadRequest(w, fmt.Errorf("mehm specification went wrong"))
		return
	}

	adminString := "false"

	if user.Admin {
		adminString = "true"
	}

	pr, err := http.NewRequest("POST", mehms+"/mehms/"+id+"/remove?userId="+user.Id+"&admin="+adminString, r.Body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	pr.Header.Set("Content-Type", "application/json")
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		utils.InternalServerError(w, err)
	}
}
