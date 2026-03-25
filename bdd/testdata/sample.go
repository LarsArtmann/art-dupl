package testdata

func Hello() {
	println("hello world")
}

type User struct {
	ID    int
	Name  string
	Email string
}

type Order struct {
	ID      int
	Product string
	Amount  float64
}

type Product struct {
	ID    int
	Name  string
	Price float64
}

type UserService struct {
	db interface{}
}

func NewUserService(db interface{}) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetByID(ctx context.Context, id int) (*User, error) {
	return &User{ID: id, Name: "Test User"}, nil
}

type OrderService struct {
	db interface{}
}

func NewOrderService(db interface{}) *OrderService {
	return &OrderService{db: db}
}

func (s *OrderService) GetByID(ctx context.Context, id int) (*Order, error) {
	return &Order{ID: id, Product: "Test Product"}, nil
}

type UserClient struct {
	apiKey string
}

func NewUserClient(apiKey string) *UserClient {
	return &UserClient{apiKey: apiKey}
}

func (c *UserClient) Create(ctx context.Context, user *User) (*User, error) {
	return user, nil
}

type ProductClient struct {
	apiKey string
}

func NewProductClient(apiKey string) *ProductClient {
	return &ProductClient{apiKey: apiKey}
}

func (c *ProductClient) Create(ctx context.Context, product *Product) (*Product, error) {
	return product, nil
}

type UserHandler struct {
	service *UserService
}

func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetUser(rec interface{}, req interface{}) {}

type OrderHandler struct {
	service *OrderService
}

func NewOrderHandler(service *OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) GetOrder(rec interface{}, req interface{}) {}

type ProductHandler struct {
	service *ProductClient
}

func NewProductHandler(service *ProductClient) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) GetProduct(rec interface{}, req interface{}) {}
