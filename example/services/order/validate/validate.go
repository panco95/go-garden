package validate

type Submit struct {
	Username string `form:"username" binding:"required,max=20,min=1" `
}