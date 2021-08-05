package pkg

type CloseNotifier struct {
	done chan struct{}
}

func NewCloseNotifier() *CloseNotifier {
	return &CloseNotifier{
		done: make(chan struct{}),
	}
}

func (n *CloseNotifier) Close() error {
	close(n.done)
	return nil
}

func (n *CloseNotifier) CloseNotify() <-chan struct{} { return n.done }
