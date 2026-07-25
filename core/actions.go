package core

const (
	ActionFileNew               = "file.new"
	ActionFileOpen              = "file.open"
	ActionFileSaveAs            = "file.save_as"
	ActionFileExportSummary     = "file.export_summary"
	ActionFileExportCabrillo    = "file.export_cabrillo"
	ActionFileExportADIF        = "file.export_adif"
	ActionFileExportCSV         = "file.export_csv"
	ActionFileExportCallhistory = "file.export_callhistory"
	ActionFileOpenRules         = "file.open_rules"
	ActionFileOpenUpload        = "file.open_upload"
	ActionFileSettings          = "file.settings"
	ActionFileConfigFile        = "file.config_file"
	ActionFileQuit              = "file.quit"

	ActionEntryClear             = "entry.clear"
	ActionEntryGotoEntryField    = "entry.goto_entry_field"
	ActionEntryEditLastQSO       = "entry.edit_last_qso"
	ActionEntryRefreshPrediction = "entry.refresh_prediction"
	ActionEntrySelectBestMatch   = "entry.select_best_match"
	ActionEntryLogQSO            = "entry.log_qso"
	ActionEntryStartParrot       = "entry.start_parrot"
	ActionEntryEnableESM         = "entry.enable_esm"
	ActionEntryNextESMStep       = "entry.next_esm_step"
	ActionEntryWorkmodeSP        = "entry.workmode_sp"
	ActionEntryWorkmodeRun       = "entry.workmode_run"
	ActionEntryOfferQTC          = "entry.offer_qtc"
	ActionEntryRequestQTC        = "entry.request_qtc"
	ActionEntryToggleFocusedVFO  = "entry.toggle_focused_vfo"
	ActionEntryFocusVFO1         = "entry.focus_vfo1"
	ActionEntryFocusVFO2         = "entry.focus_vfo2"
	ActionEntryLogVFO1           = "entry.log_vfo1"
	ActionEntryLogVFO2           = "entry.log_vfo2"
	ActionEntryClearVFO1         = "entry.clear_vfo1"
	ActionEntryClearVFO2         = "entry.clear_vfo2"

	ActionRadioXITActive        = "radio.xit_active"
	ActionRadioRITActive        = "radio.rit_active"
	ActionRadioShiftFrequency   = "radio.shift_frequency"
	ActionRadioShiftXIT         = "radio.shift_xit"
	ActionRadioShiftRIT         = "radio.shift_rit"
	ActionRadioFrequencyUp      = "radio.frequency_up"
	ActionRadioFrequencyDown    = "radio.frequency_down"
	ActionRadioXITUp            = "radio.xit_up"
	ActionRadioXITDown          = "radio.xit_down"
	ActionRadioRITUp            = "radio.rit_up"
	ActionRadioRITDown          = "radio.rit_down"
	ActionIncrementalTuningUp   = "radio.incremental_tuning_up"
	ActionIncrementalTuningDown = "radio.incremental_tuning_down"
	ActionRadioMuteAudioVFO1    = "radio.mute_audio_vfo1"
	ActionRadioMuteAudioVFO2    = "radio.mute_audio_vfo2"
	ActionRadioUnmuteAudioVFO1  = "radio.unmute_audio_vfo1"
	ActionRadioUnmuteAudioVFO2  = "radio.unmute_audio_vfo2"
	ActionRadioToggleAudioVFO1  = "radio.toggle_audio_vfo1"
	ActionRadioToggleAudioVFO2  = "radio.toggle_audio_vfo2"

	ActionBandmapMark                 = "bandmap.mark"
	ActionBandmapGotoHighestValueSpot = "bandmap.goto_highest_value_spot"
	ActionBandmapGotoNearestSpot      = "bandmap.goto_nearest_spot"
	ActionBandmapGotoNextSpotUp       = "bandmap.goto_next_spot_up"
	ActionBandmapGotoNextSpotDown     = "bandmap.goto_next_spot_down"
	ActionBandmapSendSpotsToTci       = "bandmap.send_spots_to_tci"

	ActionWindowShowQSOs       = "window.show_qsos"
	ActionWindowShowQTCs       = "window.show_qtcs"
	ActionWindowShowScoreGraph = "window.show_score_graph"
	ActionWindowShowScoreTable = "window.show_score_table"
	ActionWindowShowRate       = "window.show_rate"
	ActionWindowShowSpots      = "window.show_spots"
	ActionWindowShowClock      = "window.show_clock"

	ActionHelpWiki     = "help.wiki"
	ActionHelpSponsors = "help.sponsors"
	ActionHelpAbout    = "help.about"

	ActionKeyerSendMacro1 = "keyer.send_macro_1"
	ActionKeyerSendMacro2 = "keyer.send_macro_2"
	ActionKeyerSendMacro3 = "keyer.send_macro_3"
	ActionKeyerSendMacro4 = "keyer.send_macro_4"
	ActionKeyerShiftSpeed = "keyer.shift_speed"
	ActionKeyerSpeedUp    = "keyer.speed_up"
	ActionKeyerSpeedDown  = "keyer.speed_down"
)

// Default step sizes for the relative shift actions. Keyboard shortcuts use
// these; the remote interface may override them via an explicit amount.
const (
	DefaultFrequencyShift = Frequency(10) // Hz
	DefaultXITShift       = Frequency(10) // Hz
	DefaultRITShift       = Frequency(10) // Hz
	DefaultSpeedShift     = 1             // WPM
)
