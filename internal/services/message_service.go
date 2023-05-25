package services

// OperationService api controller of produces
type MessageService interface {
	// CreateOperation(context.Context, io.ReadCloser) (string, error)
}

type messageService struct {
	// ds datastores.OperationDatastore
}

// NewOperationService get operation service instance
func NewMessageService() MessageService {
	return &messageService{}
}
