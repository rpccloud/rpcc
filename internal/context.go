package internal

import (
	"sync/atomic"
	"unsafe"
)

type Context = *ContextObject

type ContextObject struct {
	thread unsafe.Pointer
}

func (p *ContextObject) getThread() *rpcThread {
	if p == nil {
		ReportFatalError(
			NewKernelError(ErrStringUnexpectedNil),
		)
		return nil
	} else if thread := (*rpcThread)(atomic.LoadPointer(
		&p.thread,
	)); thread == nil {
		ReportFatalError(
			NewError(ErrStringRunningOutOfScope).AddDebug(GetCodePosition("", 2)),
		)
		return nil
	} else if node := thread.execReplyNode; node == nil {
		ReportFatalError(
			NewError(ErrStringRunningOutOfScope).AddDebug(GetCodePosition("", 2)),
		)
		return nil
	} else if meta := node.replyMeta; meta == nil {
		ReportFatalError(
			NewKernelError(ErrStringUnexpectedNil),
		)
		return nil
	} else if !thread.IsDebug() {
		return thread
	} else {
		codeSource := GetCodePosition("", 2)
		switch meta.GetCheck(codeSource) {
		case rpcReplyCheckStatusOK:
			return thread
		case rpcReplyCheckStatusError:
			ReportFatalError(
				NewError(ErrStringRunningOutOfScope),
			)
			return nil
		default:
			if thread.GetGoId() != CurrentGoroutineID() {
				meta.SetCheckError(codeSource)
				ReportFatalError(
					NewError(ErrStringRunningOutOfScope),
				)
				return nil
			} else {
				meta.SetCheckOK(codeSource)
				return thread
			}
		}
	}
}

func (p *ContextObject) stop() {
	if p == nil {
		ReportFatalError(NewKernelError(ErrStringUnexpectedNil))
	} else {
		atomic.StorePointer(&p.thread, nil)
	}
}

func (p *ContextObject) OK(value interface{}) Return {
	if p == nil {
		ReportFatalError(
			NewError(ErrStringUnexpectedNil).AddDebug(GetCodePosition("", 1)),
		)
		return nilReturn
	}

	if thread := p.getThread(); thread != nil {
		return thread.WriteOK(value, 2)
	}

	// do nothing, FatalReport has already run
	return nilReturn
}

func (p *ContextObject) Error(value error) Return {
	if p == nil {
		ReportFatalError(
			NewError(ErrStringUnexpectedNil).AddDebug(GetCodePosition("", 1)),
		)
		return nilReturn
	}

	thread := p.getThread()
	if err, ok := value.(Error); ok && err != nil {
		return thread.WriteError(
			err.AddDebug(thread.GetExecReplyNodeDebug()),
		)
	} else if value != nil {
		return thread.WriteError(
			NewError(value.Error()).
				AddDebug(GetCodePosition(thread.GetExecReplyNodePath(), 1)),
		)
	} else {
		return thread.WriteError(
			NewError("value is nil").
				AddDebug(GetCodePosition(thread.GetExecReplyNodePath(), 1)),
		)
	}
}
