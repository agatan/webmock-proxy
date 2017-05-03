package webmock

type Option func(*Proxy)

func BaseDir(basedir string) Option {
	return func(p *Proxy) {
		p.BaseDir = basedir
	}
}

var RecordMode Option = func(p *Proxy) {
	p.IsRecordMode = true
}

var ReplayMode Option = func(p *Proxy) {
	p.IsRecordMode = false
}

func Addr(addr string) Option {
	return func(p *Proxy) {
		p.Addr = addr
	}
}
