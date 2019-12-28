package core

import (
	"testing"

	"github.com/andig/ulm/api"
	"github.com/andig/ulm/api/mock_api"
	"github.com/golang/mock/gomock"
)

type testCharger struct {
	api.Charger
	api.ChargeController
}

type testCase struct {
	Mode                   api.ChargeMode
	MinCurrent, MaxCurrent int64
	ActualCurrent          int64
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

	lp := NewLoadPoint("lp1", testCharger{cr, cc})
	lp.GridMeter = m

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
	var lp api.LoadPoint = NewLoadPoint("lp1", c)
	_ = lp
}

func TestChargerEnableNoChange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mock_api.NewMockCharger(ctrl)
	c.EXPECT().
		Enabled().
		Return(false, nil)

	lp := NewLoadPoint("lp1", c)
	if err := lp.chargerEnable(false); err != nil {
		t.Error(err)
	}
}

func TestChargerEnableChange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mock_api.NewMockCharger(ctrl)
	c.EXPECT().
		Enabled().
		Return(false, nil)
	c.EXPECT().
		Enable(true).
		Return(nil)

	lp := NewLoadPoint("lp1", c)
	if err := lp.chargerEnable(false); err != nil {
		t.Error(err)
	}
}

func TestEVNotConnected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mock_api.NewMockCharger(ctrl)
	c.EXPECT().
		Status().
		Return(api.StatusA, nil)

	lp := NewLoadPoint("lp1", c)

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

	lp := NewLoadPoint("lp1", c)

	lp.Update()
}

func TestEVConnectedAndEnabledNowMode(t *testing.T) {
	cases := []testCase{
		testCase{api.ModeNow, 0, 0, 0, 0.0, 16},
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

func TestEVConnectedAndEnabledPMinVMode(t *testing.T) {
	cases := []testCase{
		testCase{api.ModeMinPV, 0, 0, 5, 0.0, nil},
		testCase{api.ModeMinPV, 0, 0, 10, 1150.0, 5},
		testCase{api.ModeMinPV, 0, 0, 5, -1150.0, 10},
		testCase{api.ModeMinPV, 14, 0, 5, -1150.0, 14}, // 14A > 10A
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
		testCase{api.ModePV, 14, 0, 5, -1150.0, 0}, // 14A > 10A
	}

	for _, c := range cases {
		ctrl := gomock.NewController(t)

		lp := mockedLP(ctrl, c)
		lp.Update()

		ctrl.Finish()
	}
}
