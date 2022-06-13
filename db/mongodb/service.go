package mongodb

type mongoService struct {
	servers  string
	sessions SessionPool
}
