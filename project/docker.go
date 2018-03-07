package project

// DockerCredential holds configuration for authenticating with docker registry
type DockerCredential struct {
	Username string
	Password string
	Host     string
}
