package filterresponse

import (
	"context"
    "net"
	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
)

// Rewrite is plugin to rewrite requests internally before being handled.
type FilterResponse struct {
	Next plugin.Handler
    Filters []*net.IPNet
}

// ServeDNS implements the plugin.Handler interface.
func (fr FilterResponse) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	mw := NewResponseModifier(w, &fr)
	return plugin.NextOrFailure(fr.Name(), fr.Next, ctx, mw, r)
}

// Name implements the Handler interface.
func (fr FilterResponse) Name() string { return "filterresponse" }

type ResponseModifier struct {
	dns.ResponseWriter
    *FilterResponse
}

// Returns a dns.Msg modifier that replaces CNAME on root zones with other records.
func NewResponseModifier(w dns.ResponseWriter, fr *FilterResponse) *ResponseModifier {
	return &ResponseModifier{
		ResponseWriter: w,
        FilterResponse: fr,
	}
}

// WriteMsg records the status code and calls the
// underlying ResponseWriter's WriteMsg method.
func (r *ResponseModifier) WriteMsg(res *dns.Msg) error {
	// Find and delete A/AAAA records on that zone that fall within the suggested ranges
	for i := 0; i < len(res.Answer); {
		rr := res.Answer[i]
        var removed = false
        if t, ok := rr.(*dns.AAAA); ok {
            for _, cidr := range r.FilterResponse.Filters {
                if cidr.Contains(t.AAAA) {
                    removed = true
                }
            }
		}
        if t, ok := rr.(*dns.A); ok {
            for _, cidr := range r.FilterResponse.Filters {
                if cidr.Contains(t.A) {
                    removed = true
                }
            }
		}
        if removed {
			// Remove the CNAME record
			res.Answer = append(res.Answer[:i], res.Answer[i+1:]...)
			continue
        }
		i++
	}
	return r.ResponseWriter.WriteMsg(res)
}

// Write is a wrapper that records the size of the message that gets written.
func (r *ResponseModifier) Write(buf []byte) (int, error) {
	n, err := r.ResponseWriter.Write(buf)
	return n, err
}
