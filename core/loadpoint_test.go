package core

import (
	"testing"

	"github.com/andig/ulm/api"
	"github.com/andig/ulm/api/mock_api"
	// "github.com/andig/ulm/core"
	"github.com/golang/mock/gomock"
)

type testCharger struct {
	api.Charger
	api.ChargeController
}

type testCase struct {
	Mode                   api.ChargeMode
	MinCurrent, MaxCurrent int
	ActualCurrent          int
	CurrentPower           float64
	ExpectedCurrent        interface{}
}

func mockedLP(ctrl *gomock.Controller, tc testCase) *LoadPoint {
	cr := mock_api.NewMockCharger(ctrl)
	cr.EXPECT().
		Status().
		Return(api.StatusC, nil)
	cr.EXPECT().
		Enabled().
		Return(true, nil)
	cr.EXPECT().
		ActualCurrent().
		Return(tc.ActualCurrent, nil)

	m := mock_api.NewMockMeter(ctrl)
	m.EXPECT().
		CurrentPower().
		Return(tc.CurrentPower, nil)

	cc := mock_api.NewMockChargeController(ctrl)
	if expectedCurrent, ok := tc.ExpectedCurrent.(int); ok {
		cc.EXPECT().
			MaxCurrent(gomock.Eq(expectedCurrent)).
			Return(nil)
	}

	lp := NewLoadPoint("lp1", testCharger{cr, cc}, m)
	lp.Mode = tc.Mode
	if tc.MinCurrent > 0 {
		lp.MinCurrent = tc.MinCurrent
	}
	if tc.MaxCurrent > 0 {
		lp.MaxCurrent = tc.MaxCurrent
	}

	return lp
}

func TestNewLoadPoint(t *testing.T) {
	var c api.Charger
	var m api.Meter
	var lp api.LoadPoint = NewLoadPoint("foo", c, m)
	_ = lp
}

func TestEVNotConnected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mock_api.NewMockCharger(ctrl)
	c.EXPECT().
		Status().
		Return(api.StatusA, nil)

	lp := NewLoadPoint("lp1", c, nil)

	lp.Update()
}

func TestEVConnectedButDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mock_api.NewMockCharger(ctrl)
	c.EXPECT().
		Status().
		Return(api.StatusC, nil)
	c.EXPECT().
		Enabled().
		Return(false, nil)

	lp := NewLoadPoint("lp1", c, nil)

	lp.Update()
}

func TestEVConnectedAndEnabledNowMode(t *testing.T) {
	cases := []testCase{
		testCase{api.ModeNow, 0, 32, 0, 0.0, 32},
		testCase{api.ModeNow, 16, 32, 0, 0.0, 32},
	}

	for _, c := range cases {
		ctrl := gomock.NewController(t)

		lp := mockedLP(ctrl, c)
		lp.Update()

		ctrl.Finish()
	}
}

func TestEVConnectedAndEnabledPVMode(t *testing.T) {
	cases := []testCase{
		testCase{api.ModePV, 0, 0, 5, 0.0, nil},
		testCase{api.ModePV, 0, 0, 10, 1150.0, 5},
		testCase{api.ModePV, 0, 0, 5, -1150.0, 10},
		testCase{api.ModePV, 20, 0, 5, -1150.0, 0},
	}

	for _, c := range cases {
		ctrl := gomock.NewController(t)

		lp := mockedLP(ctrl, c)
		lp.Update()

		ctrl.Finish()
	}
}
