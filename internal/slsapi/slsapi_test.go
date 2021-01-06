// Copyright (c) 2019 Cray Inc. All Rights Reserved.
//
// Except as permitted by contract or express written permission of Cray Inc.,
// no part of this work or its content may be modified, used, reproduced or
// disclosed in any form. Modifications made without express permission of
// Cray Inc. may damage the system the software is installed within, may
// disqualify the user from receiving support from Cray Inc. under support or
// maintenance contracts, or require additional support services outside the
// scope of those contracts to repair the software or system.
//
// This file contains unit tests for the SLS API

package slsapi

///////////////////////////////////////////////////////////////////////////////
// Pre-Test Setup
///////////////////////////////////////////////////////////////////////////////

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"
)

var client *http.Client

// RoundTrip method override
type RTFunc func(req *http.Request) *http.Response

// Implement RoundTrip interface by implementing RoundTrip method
func (f RTFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(f RTFunc) *http.Client {
	return &http.Client{
		Transport: RTFunc(f),
	}
}

// Sets up the http client for testing
func TestMain(m *testing.M) {
	client = NewTestClient(NewRTFuncSLSAPI())

	excode := 1
	excode = m.Run()
	os.Exit(excode)
}

///////////////////////////////////////////////////////////////////////////////
// Unit Tests
///////////////////////////////////////////////////////////////////////////////

func TestNewSLS(t *testing.T) {
	tests := []struct {
		slsUrlIn    string
		expectedUrl string
	}{{
		slsUrlIn:    "http://cray-sls",
		expectedUrl: "http://cray-sls",
	}, {
		slsUrlIn:    "cray-sls",
		expectedUrl: "http://cray-sls",
	}}
	for i, test := range tests {
		out := NewSLS(test.slsUrlIn, nil)
		if test.expectedUrl != out.Url.String() {
			t.Errorf("Test %v Failed: Expected SLS URL '%v'; Received SLS URL '%v'", i, test.expectedUrl, out.Url.String())
		}
	}
}

func TestIsReady(t *testing.T) {
	defaultUrl, _ := url.Parse("http://cray-sls")
	badUrl, _ := url.Parse("http://cray-sls.bad")
	tests := []struct {
		SLSUrl      *url.URL
		Client      *http.Client
		expectedRsp bool
		expectErr   bool
	}{{
		SLSUrl:      defaultUrl,
		Client:      client,
		expectedRsp: true,
		expectErr:   false,
	}, {
		SLSUrl:      badUrl,
		Client:      client,
		expectedRsp: false,
		expectErr:   true,
	}}

	for i, test := range tests {
		sls := SLS{
			Url: test.SLSUrl,
			Client: test.Client,
		}
		out, err := sls.IsReady()
		if err != nil {
			if !test.expectErr {
				t.Errorf("Test %v Failed: Unexpected error - %v", i, err)
			}
		} else {
			if test.expectErr {
				t.Errorf("Test %v Failed: Expected an error", i)
			} else {
				if out != test.expectedRsp {
					t.Errorf("Test %v Failed: Expected status '%v'; Received status '%v'", i, test.expectedRsp, out)
				}
			}
		}
	}
}

func TestGetNodeInfo(t *testing.T) {
	defaultUrl, _ := url.Parse("http://cray-sls")
	tests := []struct {
		SLSUrl       *url.URL
		Client       *http.Client
		id           string
		expectedInfo ComptypeNode
		expectErr    bool
	}{{
		SLSUrl:       defaultUrl,
		Client:       client,
		id:           "x0c0s0b0n0",
		expectedInfo: ComptypeNode{
			NID:  1,
			Role: "Application",
		},
		expectErr:    false,
	}, {
		SLSUrl:       defaultUrl,
		Client:       client,
		id:           "x0c0s1b0n0",
		expectedInfo: ComptypeNode{
			Role: "Application",
		},
		expectErr:    false,
	}, {
		SLSUrl:       defaultUrl,
		Client:       client,
		id:           "x0c0s2b0n0",
		expectedInfo: ComptypeNode{
			NID:  1,
		},
		expectErr:    false,
	}, {
		SLSUrl:       defaultUrl,
		Client:       client,
		id:           "x0c0s3b0n0",
		expectedInfo: ComptypeNode{},
		expectErr:    false,
	}, {
		SLSUrl:       defaultUrl,
		Client:       client,
		id:           "x0c0s4b0n0",
		expectedInfo: ComptypeNode{},
		expectErr:    false,
	}, {
		SLSUrl:       nil,
		Client:       client,
		id:           "x0c0s4b0n0",
		expectedInfo: ComptypeNode{},
		expectErr:    true,
	}, {
		SLSUrl:       defaultUrl,
		Client:       client,
		id:           "",
		expectedInfo: ComptypeNode{},
		expectErr:    true,
	}, {
		SLSUrl:       defaultUrl,
		Client:       nil,
		id:           "x0c0s4b0n0",
		expectedInfo: ComptypeNode{},
		expectErr:    true,
	}}

	for i, test := range tests {
		sls := SLS{
			Url: test.SLSUrl,
			Client: test.Client,
		}
		out, err := sls.GetNodeInfo(test.id)
		if err != nil {
			if !test.expectErr {
				t.Errorf("Test %v Failed: Unexpected error - %v", i, err)
			}
		} else {
			if test.expectErr {
				t.Errorf("Test %v Failed: Expected an error", i)
			} else {
				if out.NID != test.expectedInfo.NID &&
					out.Role != test.expectedInfo.Role {
					t.Errorf("Test %v Failed: Expected node info '%v'; Received node info '%v'", i, test.expectedInfo, out)
				}
			}
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// Mock SLS Data
///////////////////////////////////////////////////////////////////////////////

const testPayloadSLSAPI_ready = `
{
	"ready": true,
	"code": 0
}`

const testPayloadSLSAPI_notReady = `
{
	"ready": false,
	"code": 503,
	"reason": "Things are pretty bad"
}`

const testPayloadSLSAPI_goodComp = `
{
	"Class": "River",
	"ExtraProperties": {
		"Role": "Application",
		"NID": 1
	},
	"Parent": "x0c0s0b0",
	"Type": "comptype_node",
	"TypeString": "Node",
	"Xname": "x0c0s0b0n0"
}`

const testPayloadSLSAPI_comp_noNID = `
{
	"Class": "River",
	"ExtraProperties": {
		"Role": "Application"
	},
	"Parent": "x0c0s1b0",
	"Type": "comptype_node",
	"TypeString": "Node",
	"Xname": "x0c0s1b0n0"
}`

const testPayloadSLSAPI_comp_noRole = `
{
	"Class": "River",
	"ExtraProperties": {
		"Role": "",
		"NID": 1
	},
	"Parent": "x0c0s2b0",
	"Type": "comptype_node",
	"TypeString": "Node",
	"Xname": "x0c0s2b0n0"
}`

const testPayloadSLSAPI_comp_noData = `
{
	"Class": "River",
	"Parent": "x0c0s3b0",
	"Type": "comptype_node",
	"TypeString": "Node",
	"Xname": "x0c0s3b0n0"
}`

func NewRTFuncSLSAPI() RTFunc {
	return func(req *http.Request) *http.Response {
		// Test request parameters
		switch req.URL.String() {
		case "http://cray-sls/ready":
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadSLSAPI_ready)),
				Header: make(http.Header),
			}
		case "http://cray-sls.bad/ready":
			return &http.Response{
				StatusCode: 503,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadSLSAPI_notReady)),
				Header: make(http.Header),
			}
		case "http://cray-sls/hardware/x0c0s0b0n0":
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadSLSAPI_goodComp)),
				Header: make(http.Header),
			}
		case "http://cray-sls/hardware/x0c0s1b0n0":
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadSLSAPI_comp_noNID)),
				Header: make(http.Header),
			}
		case "http://cray-sls/hardware/x0c0s2b0n0":
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadSLSAPI_comp_noRole)),
				Header: make(http.Header),
			}
		case "http://cray-sls/hardware/x0c0s3b0n0":
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadSLSAPI_comp_noData)),
				Header: make(http.Header),
			}
		case "http://cray-sls/hardware/x0c0s4b0n0":
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString("null")),
				Header: make(http.Header),
			}
		default:
			return &http.Response{
				StatusCode: 404,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString("")),
				Header: make(http.Header),
			}
		}
	}
}