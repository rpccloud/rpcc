package internal

type rpcReplyMeta struct {
	name    string      // the name of reply
	handler interface{} // reply handler
	debug   string      // where the reply add in source file
}

type rpcAddChildMeta struct {
	name    string   // the name of child service
	service *Service // the real service
	debug   string   // where the service add in source file
}

type Service struct {
	children []*rpcAddChildMeta // all the children node meta pointer
	replies  []*rpcReplyMeta    // all the replies meta pointer
	debug    string             // where the service define in source file
	Lock
}

// NewService define a new service
func NewService() *Service {
	return &Service{
		children: make([]*rpcAddChildMeta, 0, 0),
		replies:  make([]*rpcReplyMeta, 0, 0),
		debug:    GetStackString(1),
	}
}

// Reply add reply handler
func (p *Service) Reply(
	name string,
	handler interface{},
) *Service {
	p.DoWithLock(func() {
		// add reply meta
		p.replies = append(p.replies, &rpcReplyMeta{
			name:    name,
			handler: handler,
			debug:   GetStackString(3),
		})
	})
	return p
}

// AddChild add child service
func (p *Service) AddChild(name string, service *Service) *Service {
	// lock
	p.DoWithLock(func() {
		// add child meta
		p.children = append(p.children, &rpcAddChildMeta{
			name:    name,
			service: service,
			debug:   GetStackString(3),
		})
	})

	return p
}
