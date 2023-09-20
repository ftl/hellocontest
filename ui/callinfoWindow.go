package ui

import (
	"github.com/ftl/gmtry"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui/style"
)

const CallinfoWindowID = gmtry.ID("callinfo")

type callinfoWindow struct {
	callinfoView *callinfoView
	controller   CallinfoController

	window   *gtk.Window
	geometry *gmtry.Geometry
	style    *style.Style

	callsign          string
	worked            bool
	duplicate         bool
	dxccName          string
	continent         string
	itu               int
	cq                int
	arrlCompliant     bool
	points            int
	multis            int
	exchange          string
	userInfo          string
	matchingCallsigns []core.AnnotatedCallsign
}

func setupCallinfoWindow(geometry *gmtry.Geometry, style *style.Style, controller CallinfoController) *callinfoWindow {
	result := &callinfoWindow{
		controller: controller,
		geometry:   geometry,
		style:      style,
	}

	return result
}

func (w *callinfoWindow) RestoreVisibility() {
	visible := w.geometry.Get(CallinfoWindowID).Visible
	if visible {
		w.Show()
	} else {
		w.Hide()
	}
}

func (w *callinfoWindow) Show() {
	if w.window == nil {
		builder := setupBuilder()
		w.window = getUI(builder, "callinfoWindow").(*gtk.Window)
		w.window.SetDefaultSize(300, 500)
		w.window.SetTitle("Callsign Information")
		w.window.Connect("destroy", w.onDestroy)
		w.callinfoView = setupCallinfoView(builder, w.style.ForWidget(w.window.ToWidget()), w.controller)
		w.callinfoView.SetCallsign(w.callsign, w.worked, w.duplicate)
		w.callinfoView.SetDXCC(w.dxccName, w.continent, w.itu, w.cq, w.arrlCompliant)
		w.callinfoView.SetValue(w.points, w.multis)
		w.callinfoView.SetExchange(w.exchange)
		w.callinfoView.SetUserInfo(w.userInfo)
		w.callinfoView.SetSupercheck(w.matchingCallsigns)
		w.window.Connect("style-updated", w.callinfoView.RefreshStyle)
		connectToGeometry(w.geometry, CallinfoWindowID, w.window)
	}
	w.window.ShowAll()
	w.window.Present()
}

func (w *callinfoWindow) Hide() {
	if w.window == nil {
		return
	}
	w.window.Close()
}

func (w *callinfoWindow) Visible() bool {
	if w.window == nil {
		return false
	}
	return w.window.IsVisible()
}

func (w *callinfoWindow) UseDefaultWindowGeometry() {
	if w.window == nil {
		return
	}
	w.window.Move(0, 100)
	w.window.Resize(300, 500)
}

func (w *callinfoWindow) onDestroy() {
	w.window = nil
	w.callinfoView = nil
}

func (w *callinfoWindow) SetCallsign(callsign string, worked, duplicate bool) {
	w.callsign = callsign
	w.worked = worked
	w.duplicate = duplicate

	if w.callinfoView != nil {
		w.callinfoView.SetCallsign(callsign, worked, duplicate)
	}
}

func (w *callinfoWindow) SetDXCC(dxccName, continent string, itu, cq int, arrlCompliant bool) {
	w.dxccName = dxccName
	w.continent = continent
	w.itu = itu
	w.cq = cq
	w.arrlCompliant = arrlCompliant

	if w.callinfoView != nil {
		w.callinfoView.SetDXCC(dxccName, continent, itu, cq, arrlCompliant)
	}
}

func (w *callinfoWindow) SetValue(points, multis int) {
	w.points = points
	w.multis = multis

	if w.callinfoView != nil {
		w.callinfoView.SetValue(points, multis)
	}
}

func (w *callinfoWindow) SetExchange(exchange string) {
	w.exchange = exchange

	if w.callinfoView != nil {
		w.callinfoView.SetExchange(exchange)
	}
}

func (w *callinfoWindow) SetUserInfo(userInfo string) {
	w.userInfo = userInfo

	if w.callinfoView != nil {
		w.callinfoView.SetUserInfo(userInfo)
	}
}

func (w *callinfoWindow) SetSupercheck(matchingCallsigns []core.AnnotatedCallsign) {
	w.matchingCallsigns = matchingCallsigns

	if w.callinfoView != nil {
		w.callinfoView.SetSupercheck(matchingCallsigns)
	}
}
