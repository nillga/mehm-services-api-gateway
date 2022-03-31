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
var users = os.Getenv("USERS_HOST")
var mehms = os.Getenv("MEHMS_HOST")

func NewApiGatewayController() ApiGatewayController {
	return &controller{}
}

// GetMehms godoc
// @Security bearerToken
// @Summary      Read a page of mehms
// @Description  Pagination can be handled via query parameters
// @Tags         mehms
// @Accept       json
// @Produce      json
// @Param        skip   query      int  false  "states the number of skipped Mehms" minimum(0) default(0)
// @Param        take   query      int  false  "states the count of grabbed Mehms" minimum(1) maximum(30) default(30)
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
// @Summary      View a specified mehm
// @Security bearerToken
// @Description  This will return the requested Mehm including the information whether you have liked it already.
// @Tags         mehms
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "The ID of the requested mehm" minimum(1)
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
// @Summary      Read a specified comment
// @Security bearerToken
// @Description  By specifying the comment id, you can read that comment
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "The ID of the requested comment" minimum(1)
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
// @Summary      Profile information
// @Security bearerToken
// @Description  This call will respond with your id, username and email.
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
// @Summary      Like a specified mehm
// @Security bearerToken
// @Description  This is a like-toggle: if the Mehm had been liked already, the like will be removed.
// @Tags         mehms
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "The ID of the requested mehm" minimum(1)
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
// @Summary      Post a comment
// @Security bearerToken
// @Description  With this API-Call you are able to post a comment related to any existing Mehm.
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        comment   query      string  true  "The comment" minlength(1) maxlength(256)
// @Param        mehmId   query      int  true  "The mehm" minimum(1)
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
// @Summary      Edit an existing comment
// @Security bearerToken
// @Description  Here you can edit previously posted comments. An Admin will be able to edit other people's comments too.
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
// @Summary      Edit a Mehm's shown information
// @Security bearerToken
// @Description  This will be only possible for own Mehms, unless you are privileged
// @Tags         mehms
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "The ID of the requested mehm" minimum(1)
// @Param		 description body string true "The new mehm description" minlength(1) maxlength(128)
// @Param        title body string true "The new mehm title" minlength(1) maxlength(32)
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
// @Summary      Delete a user
// @Security bearerToken
// @Description  Regular user can only delete theirselves, admin users can delete every user
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
// @Summary      Delete a Mehm
// @Security bearerToken
// @Description  Regular users can only delete their own Mehms, privileged users can delete whatever they wish
// @Tags         mehms
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "The ID of the requested mehm" minimum(1)
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
