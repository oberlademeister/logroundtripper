package logroundtripper

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// LogRoundTripper is a golang RoundTripper object which encapsulates the std library http.Transport
// but - for each request, writes a log file entry to an io.Writer
type LogRoundTripper struct {
	Transport *http.Transport
	DryRun    bool
	Out       io.Writer
	requestID int16
}

type readCloser struct {
	io.Reader
}

func (r *readCloser) Close() error {
	return nil
}

// RoundTrip implements the RoundTripper Interface
func (lrt *LogRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	thisID := lrt.requestID
	lrt.requestID++
	t0 := time.Now()
	{
		var reqBuf bytes.Buffer
		var reqBody []byte
		fmt.Fprintf(&reqBuf, "--\nREQ Time: %d.%03d ID-%04x\n", t0.Unix(), t0.UnixNano()%1000, thisID)
		fmt.Fprintf(&reqBuf, "%s %q\n", req.Method, req.URL)
		if req.Body != nil {
			var err error
			reqBody, err = ioutil.ReadAll(req.Body)
			req.Body.Close()
			if err != nil {
				return nil, err
			}
			fmt.Fprintf(&reqBuf, "%s\n", string(reqBody))
			req.Body = &readCloser{bytes.NewReader(reqBody)}
		}
		// now log request piece
		lrt.Out.Write(reqBuf.Bytes())
	}
	{
		if lrt.DryRun {
			return &http.Response{}, nil
		}
		// now to the response
		resp, err := lrt.Transport.RoundTrip(req)
		if err != nil {
			return nil, err
		}
		var resBuf bytes.Buffer
		var resBody []byte
		t1 := time.Now()
		fmt.Fprintf(&resBuf, "--\nRES Time: %d.%03d %dms ID-%04x Status: %d\n", t1.Unix(), t1.UnixNano()%1000, t1.Sub(t0).Nanoseconds()/1000000, thisID, resp.StatusCode)
		if resp.Body != nil {
			var err error
			resBody, err = ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return nil, err
			}
			if len(resBody) > 0 {
				fmt.Fprintf(&resBuf, "%s\n", string(resBody))
			}
			resp.Body = &readCloser{bytes.NewReader(resBody)}
		}
		lrt.Out.Write(resBuf.Bytes())
		return resp, err
	}
}
