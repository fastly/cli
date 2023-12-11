package testutil

import (
	"github.com/fastly/go-fastly/v8/fastly"
)

// ServicesPaginator mocks the behaviour of a paginator for services.
type ServicesPaginator struct {
	Count         int
	MaxPages      int
	NumOfPages    int
	RequestedPage int
	ReturnErr     bool
}

// HasNext indicates if there is another page of data.
func (p *ServicesPaginator) HasNext() bool {
	if p.Count > p.MaxPages {
		return false
	}
	p.Count++
	return true
}

// Remaining returns the count of remaining pages.
func (p ServicesPaginator) Remaining() int {
	return 1
}

// GetNext returns the next page of data.
func (p *ServicesPaginator) GetNext() (ss []*fastly.Service, err error) {
	if p.ReturnErr {
		err = Err
	}
	pageOne := fastly.Service{
		ID:            fastly.ToPointer("123"),
		Name:          fastly.ToPointer("Foo"),
		Type:          fastly.ToPointer("wasm"),
		CustomerID:    fastly.ToPointer("mycustomerid"),
		ActiveVersion: fastly.ToPointer(2),
		UpdatedAt:     MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
		Versions: []*fastly.Version{
			{
				Number:    fastly.ToPointer(1),
				Comment:   fastly.ToPointer("a"),
				ServiceID: fastly.ToPointer("b"),
				CreatedAt: MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
				UpdatedAt: MustParseTimeRFC3339("2001-02-04T04:05:06Z"),
				DeletedAt: MustParseTimeRFC3339("2001-02-05T04:05:06Z"),
			},
			{
				Number:    fastly.ToPointer(2),
				Comment:   fastly.ToPointer("c"),
				ServiceID: fastly.ToPointer("d"),
				Active:    fastly.ToPointer(true),
				Deployed:  fastly.ToPointer(true),
				CreatedAt: MustParseTimeRFC3339("2001-03-03T04:05:06Z"),
				UpdatedAt: MustParseTimeRFC3339("2001-03-04T04:05:06Z"),
			},
		},
	}
	pageTwo := fastly.Service{
		ID:            fastly.ToPointer("456"),
		Name:          fastly.ToPointer("Bar"),
		Type:          fastly.ToPointer("wasm"),
		CustomerID:    fastly.ToPointer("mycustomerid"),
		ActiveVersion: fastly.ToPointer(1),
		UpdatedAt:     MustParseTimeRFC3339("2015-03-14T12:59:59Z"),
	}
	pageThree := fastly.Service{
		ID:            fastly.ToPointer("789"),
		Name:          fastly.ToPointer("Baz"),
		Type:          fastly.ToPointer("vcl"),
		CustomerID:    fastly.ToPointer("mycustomerid"),
		ActiveVersion: fastly.ToPointer(1),
	}
	if p.Count == 1 {
		ss = append(ss, &pageOne)
	}
	if p.Count == 2 {
		ss = append(ss, &pageTwo)
	}
	if p.Count == 3 {
		ss = append(ss, &pageThree)
	}
	if p.RequestedPage > 0 && p.NumOfPages == 1 {
		p.Count = p.MaxPages + 1 // forces only one result to be displayed
	}
	return ss, err
}
