package repository

type RepositoryResponse struct {
	Success        bool
	SuccessMessage string
	Place          string
	Data           map[string]any
	Errors         map[string]string
}

const LoadPhoto = "Repository-LoadPhoto"
const DeletePhoto = "Repository-DeletePhoto"
const GetPhotos = "Repository-GetPhotos"
const GetPhoto = "Repository-GetPhoto"
const DeleteUserData = "Repository-DeleteUserData"
const AddUserId = "Repository-AddUserId"
const KeyUsersIdTable = "usersid"
const KeyPhotoTable = "photos"
const KeyPhotoID = "photoid"
const KeyUserID = "userid"
const KeyPhotoURL = "url"
const KeyPhotoSize = "size"
const KeyContentType = "content_type"
const KeyCreatedTime = "created_at"
const KeyPhoto = "photo"
