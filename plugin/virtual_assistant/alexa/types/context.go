package types

type Context struct {
	Properties []Property `json:"properties"`
}

type Property struct {
	Namespace                 string      `json:"namespace"`
	Instance                  string      `json:"instance,omitempty"`
	Name                      string      `json:"name"`
	Value                     interface{} `json:"value"`
	TimeOfSample              string      `json:"timeOfSample"`
	UncertaintyInMilliseconds uint        `json:"uncertaintyInMilliseconds"`
}
