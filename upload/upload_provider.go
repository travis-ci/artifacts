package upload

type uploadProvider interface {
	Upload(string, *Options, chan *artifact, chan *artifact, chan bool)
	Name() string
}
