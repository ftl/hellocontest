// Code generated by protoc-gen-go. DO NOT EDIT.
// source: log.proto

package pb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type FileInfo struct {
	FormatVersion        int32    `protobuf:"varint,1,opt,name=format_version,json=formatVersion,proto3" json:"format_version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FileInfo) Reset()         { *m = FileInfo{} }
func (m *FileInfo) String() string { return proto.CompactTextString(m) }
func (*FileInfo) ProtoMessage()    {}
func (*FileInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_a153da538f858886, []int{0}
}

func (m *FileInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FileInfo.Unmarshal(m, b)
}
func (m *FileInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FileInfo.Marshal(b, m, deterministic)
}
func (m *FileInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FileInfo.Merge(m, src)
}
func (m *FileInfo) XXX_Size() int {
	return xxx_messageInfo_FileInfo.Size(m)
}
func (m *FileInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_FileInfo.DiscardUnknown(m)
}

var xxx_messageInfo_FileInfo proto.InternalMessageInfo

func (m *FileInfo) GetFormatVersion() int32 {
	if m != nil {
		return m.FormatVersion
	}
	return 0
}

type Entry struct {
	// Types that are valid to be assigned to Entry:
	//
	//	*Entry_Qso
	//	*Entry_Station
	//	*Entry_Contest
	//	*Entry_Keyer
	Entry                isEntry_Entry `protobuf_oneof:"entry"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *Entry) Reset()         { *m = Entry{} }
func (m *Entry) String() string { return proto.CompactTextString(m) }
func (*Entry) ProtoMessage()    {}
func (*Entry) Descriptor() ([]byte, []int) {
	return fileDescriptor_a153da538f858886, []int{1}
}

func (m *Entry) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Entry.Unmarshal(m, b)
}
func (m *Entry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Entry.Marshal(b, m, deterministic)
}
func (m *Entry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Entry.Merge(m, src)
}
func (m *Entry) XXX_Size() int {
	return xxx_messageInfo_Entry.Size(m)
}
func (m *Entry) XXX_DiscardUnknown() {
	xxx_messageInfo_Entry.DiscardUnknown(m)
}

var xxx_messageInfo_Entry proto.InternalMessageInfo

type isEntry_Entry interface {
	isEntry_Entry()
}

type Entry_Qso struct {
	Qso *QSO `protobuf:"bytes,1,opt,name=qso,proto3,oneof"`
}

type Entry_Station struct {
	Station *Station `protobuf:"bytes,2,opt,name=station,proto3,oneof"`
}

type Entry_Contest struct {
	Contest *Contest `protobuf:"bytes,3,opt,name=contest,proto3,oneof"`
}

type Entry_Keyer struct {
	Keyer *Keyer `protobuf:"bytes,4,opt,name=keyer,proto3,oneof"`
}

func (*Entry_Qso) isEntry_Entry() {}

func (*Entry_Station) isEntry_Entry() {}

func (*Entry_Contest) isEntry_Entry() {}

func (*Entry_Keyer) isEntry_Entry() {}

func (m *Entry) GetEntry() isEntry_Entry {
	if m != nil {
		return m.Entry
	}
	return nil
}

func (m *Entry) GetQso() *QSO {
	if x, ok := m.GetEntry().(*Entry_Qso); ok {
		return x.Qso
	}
	return nil
}

func (m *Entry) GetStation() *Station {
	if x, ok := m.GetEntry().(*Entry_Station); ok {
		return x.Station
	}
	return nil
}

func (m *Entry) GetContest() *Contest {
	if x, ok := m.GetEntry().(*Entry_Contest); ok {
		return x.Contest
	}
	return nil
}

func (m *Entry) GetKeyer() *Keyer {
	if x, ok := m.GetEntry().(*Entry_Keyer); ok {
		return x.Keyer
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Entry) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Entry_Qso)(nil),
		(*Entry_Station)(nil),
		(*Entry_Contest)(nil),
		(*Entry_Keyer)(nil),
	}
}

type QSO struct {
	Callsign             string   `protobuf:"bytes,1,opt,name=callsign,proto3" json:"callsign,omitempty"`
	Timestamp            int64    `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Band                 string   `protobuf:"bytes,3,opt,name=band,proto3" json:"band,omitempty"`
	Mode                 string   `protobuf:"bytes,4,opt,name=mode,proto3" json:"mode,omitempty"`
	MyReport             string   `protobuf:"bytes,5,opt,name=my_report,json=myReport,proto3" json:"my_report,omitempty"`
	MyNumber             int32    `protobuf:"varint,6,opt,name=my_number,json=myNumber,proto3" json:"my_number,omitempty"`
	TheirReport          string   `protobuf:"bytes,7,opt,name=their_report,json=theirReport,proto3" json:"their_report,omitempty"`
	TheirNumber          int32    `protobuf:"varint,8,opt,name=their_number,json=theirNumber,proto3" json:"their_number,omitempty"`
	LogTimestamp         int64    `protobuf:"varint,9,opt,name=log_timestamp,json=logTimestamp,proto3" json:"log_timestamp,omitempty"`
	MyXchange            string   `protobuf:"bytes,10,opt,name=my_xchange,json=myXchange,proto3" json:"my_xchange,omitempty"`
	TheirXchange         string   `protobuf:"bytes,11,opt,name=their_xchange,json=theirXchange,proto3" json:"their_xchange,omitempty"`
	Frequency            float64  `protobuf:"fixed64,12,opt,name=frequency,proto3" json:"frequency,omitempty"`
	MyExchange           []string `protobuf:"bytes,14,rep,name=my_exchange,json=myExchange,proto3" json:"my_exchange,omitempty"`
	TheirExchange        []string `protobuf:"bytes,15,rep,name=their_exchange,json=theirExchange,proto3" json:"their_exchange,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *QSO) Reset()         { *m = QSO{} }
func (m *QSO) String() string { return proto.CompactTextString(m) }
func (*QSO) ProtoMessage()    {}
func (*QSO) Descriptor() ([]byte, []int) {
	return fileDescriptor_a153da538f858886, []int{2}
}

func (m *QSO) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QSO.Unmarshal(m, b)
}
func (m *QSO) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QSO.Marshal(b, m, deterministic)
}
func (m *QSO) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QSO.Merge(m, src)
}
func (m *QSO) XXX_Size() int {
	return xxx_messageInfo_QSO.Size(m)
}
func (m *QSO) XXX_DiscardUnknown() {
	xxx_messageInfo_QSO.DiscardUnknown(m)
}

var xxx_messageInfo_QSO proto.InternalMessageInfo

func (m *QSO) GetCallsign() string {
	if m != nil {
		return m.Callsign
	}
	return ""
}

func (m *QSO) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *QSO) GetBand() string {
	if m != nil {
		return m.Band
	}
	return ""
}

func (m *QSO) GetMode() string {
	if m != nil {
		return m.Mode
	}
	return ""
}

func (m *QSO) GetMyReport() string {
	if m != nil {
		return m.MyReport
	}
	return ""
}

func (m *QSO) GetMyNumber() int32 {
	if m != nil {
		return m.MyNumber
	}
	return 0
}

func (m *QSO) GetTheirReport() string {
	if m != nil {
		return m.TheirReport
	}
	return ""
}

func (m *QSO) GetTheirNumber() int32 {
	if m != nil {
		return m.TheirNumber
	}
	return 0
}

func (m *QSO) GetLogTimestamp() int64 {
	if m != nil {
		return m.LogTimestamp
	}
	return 0
}

func (m *QSO) GetMyXchange() string {
	if m != nil {
		return m.MyXchange
	}
	return ""
}

func (m *QSO) GetTheirXchange() string {
	if m != nil {
		return m.TheirXchange
	}
	return ""
}

func (m *QSO) GetFrequency() float64 {
	if m != nil {
		return m.Frequency
	}
	return 0
}

func (m *QSO) GetMyExchange() []string {
	if m != nil {
		return m.MyExchange
	}
	return nil
}

func (m *QSO) GetTheirExchange() []string {
	if m != nil {
		return m.TheirExchange
	}
	return nil
}

type Station struct {
	Callsign             string   `protobuf:"bytes,1,opt,name=callsign,proto3" json:"callsign,omitempty"`
	Operator             string   `protobuf:"bytes,2,opt,name=operator,proto3" json:"operator,omitempty"`
	Locator              string   `protobuf:"bytes,3,opt,name=locator,proto3" json:"locator,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Station) Reset()         { *m = Station{} }
func (m *Station) String() string { return proto.CompactTextString(m) }
func (*Station) ProtoMessage()    {}
func (*Station) Descriptor() ([]byte, []int) {
	return fileDescriptor_a153da538f858886, []int{3}
}

func (m *Station) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Station.Unmarshal(m, b)
}
func (m *Station) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Station.Marshal(b, m, deterministic)
}
func (m *Station) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Station.Merge(m, src)
}
func (m *Station) XXX_Size() int {
	return xxx_messageInfo_Station.Size(m)
}
func (m *Station) XXX_DiscardUnknown() {
	xxx_messageInfo_Station.DiscardUnknown(m)
}

var xxx_messageInfo_Station proto.InternalMessageInfo

func (m *Station) GetCallsign() string {
	if m != nil {
		return m.Callsign
	}
	return ""
}

func (m *Station) GetOperator() string {
	if m != nil {
		return m.Operator
	}
	return ""
}

func (m *Station) GetLocator() string {
	if m != nil {
		return m.Locator
	}
	return ""
}

type Contest struct {
	Name                    string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	EnterTheirNumber        bool                   `protobuf:"varint,2,opt,name=enter_their_number,json=enterTheirNumber,proto3" json:"enter_their_number,omitempty"`
	EnterTheirXchange       bool                   `protobuf:"varint,3,opt,name=enter_their_xchange,json=enterTheirXchange,proto3" json:"enter_their_xchange,omitempty"`
	RequireTheirXchange     bool                   `protobuf:"varint,4,opt,name=require_their_xchange,json=requireTheirXchange,proto3" json:"require_their_xchange,omitempty"`
	AllowMultiBand          bool                   `protobuf:"varint,5,opt,name=allow_multi_band,json=allowMultiBand,proto3" json:"allow_multi_band,omitempty"`
	AllowMultiMode          bool                   `protobuf:"varint,6,opt,name=allow_multi_mode,json=allowMultiMode,proto3" json:"allow_multi_mode,omitempty"`
	SameCountryPoints       int32                  `protobuf:"varint,7,opt,name=same_country_points,json=sameCountryPoints,proto3" json:"same_country_points,omitempty"`
	SameContinentPoints     int32                  `protobuf:"varint,8,opt,name=same_continent_points,json=sameContinentPoints,proto3" json:"same_continent_points,omitempty"`
	SpecificCountryPoints   int32                  `protobuf:"varint,9,opt,name=specific_country_points,json=specificCountryPoints,proto3" json:"specific_country_points,omitempty"`
	SpecificCountryPrefixes []string               `protobuf:"bytes,10,rep,name=specific_country_prefixes,json=specificCountryPrefixes,proto3" json:"specific_country_prefixes,omitempty"`
	OtherPoints             int32                  `protobuf:"varint,11,opt,name=other_points,json=otherPoints,proto3" json:"other_points,omitempty"`
	Multis                  *Multis                `protobuf:"bytes,12,opt,name=multis,proto3" json:"multis,omitempty"`
	XchangeMultiPattern     string                 `protobuf:"bytes,13,opt,name=xchange_multi_pattern,json=xchangeMultiPattern,proto3" json:"xchange_multi_pattern,omitempty"`
	CountPerBand            bool                   `protobuf:"varint,14,opt,name=count_per_band,json=countPerBand,proto3" json:"count_per_band,omitempty"`
	CabrilloQsoTemplate     string                 `protobuf:"bytes,15,opt,name=cabrillo_qso_template,json=cabrilloQsoTemplate,proto3" json:"cabrillo_qso_template,omitempty"`
	CallHistoryFilename     string                 `protobuf:"bytes,16,opt,name=call_history_filename,json=callHistoryFilename,proto3" json:"call_history_filename,omitempty"`
	DefinitionYaml          string                 `protobuf:"bytes,18,opt,name=definition_yaml,json=definitionYaml,proto3" json:"definition_yaml,omitempty"`
	ExchangeValues          []string               `protobuf:"bytes,19,rep,name=exchange_values,json=exchangeValues,proto3" json:"exchange_values,omitempty"`
	GenerateSerialExchange  bool                   `protobuf:"varint,20,opt,name=generate_serial_exchange,json=generateSerialExchange,proto3" json:"generate_serial_exchange,omitempty"`
	CallHistoryFieldNames   []string               `protobuf:"bytes,21,rep,name=call_history_field_names,json=callHistoryFieldNames,proto3" json:"call_history_field_names,omitempty"`
	QsosGoal                int32                  `protobuf:"varint,22,opt,name=qsos_goal,json=qsosGoal,proto3" json:"qsos_goal,omitempty"`
	PointsGoal              int32                  `protobuf:"varint,23,opt,name=points_goal,json=pointsGoal,proto3" json:"points_goal,omitempty"`
	MultisGoal              int32                  `protobuf:"varint,24,opt,name=multis_goal,json=multisGoal,proto3" json:"multis_goal,omitempty"`
	SprintOperation         bool                   `protobuf:"varint,25,opt,name=sprint_operation,json=sprintOperation,proto3" json:"sprint_operation,omitempty"`
	StartTime               *timestamppb.Timestamp `protobuf:"bytes,26,opt,name=start_time,json=startTime,proto3" json:"start_time,omitempty"`
	GenerateReport          bool                   `protobuf:"varint,27,opt,name=generate_report,json=generateReport,proto3" json:"generate_report,omitempty"`
	XXX_NoUnkeyedLiteral    struct{}               `json:"-"`
	XXX_unrecognized        []byte                 `json:"-"`
	XXX_sizecache           int32                  `json:"-"`
}

func (m *Contest) Reset()         { *m = Contest{} }
func (m *Contest) String() string { return proto.CompactTextString(m) }
func (*Contest) ProtoMessage()    {}
func (*Contest) Descriptor() ([]byte, []int) {
	return fileDescriptor_a153da538f858886, []int{4}
}

func (m *Contest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Contest.Unmarshal(m, b)
}
func (m *Contest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Contest.Marshal(b, m, deterministic)
}
func (m *Contest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Contest.Merge(m, src)
}
func (m *Contest) XXX_Size() int {
	return xxx_messageInfo_Contest.Size(m)
}
func (m *Contest) XXX_DiscardUnknown() {
	xxx_messageInfo_Contest.DiscardUnknown(m)
}

var xxx_messageInfo_Contest proto.InternalMessageInfo

func (m *Contest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Contest) GetEnterTheirNumber() bool {
	if m != nil {
		return m.EnterTheirNumber
	}
	return false
}

func (m *Contest) GetEnterTheirXchange() bool {
	if m != nil {
		return m.EnterTheirXchange
	}
	return false
}

func (m *Contest) GetRequireTheirXchange() bool {
	if m != nil {
		return m.RequireTheirXchange
	}
	return false
}

func (m *Contest) GetAllowMultiBand() bool {
	if m != nil {
		return m.AllowMultiBand
	}
	return false
}

func (m *Contest) GetAllowMultiMode() bool {
	if m != nil {
		return m.AllowMultiMode
	}
	return false
}

func (m *Contest) GetSameCountryPoints() int32 {
	if m != nil {
		return m.SameCountryPoints
	}
	return 0
}

func (m *Contest) GetSameContinentPoints() int32 {
	if m != nil {
		return m.SameContinentPoints
	}
	return 0
}

func (m *Contest) GetSpecificCountryPoints() int32 {
	if m != nil {
		return m.SpecificCountryPoints
	}
	return 0
}

func (m *Contest) GetSpecificCountryPrefixes() []string {
	if m != nil {
		return m.SpecificCountryPrefixes
	}
	return nil
}

func (m *Contest) GetOtherPoints() int32 {
	if m != nil {
		return m.OtherPoints
	}
	return 0
}

func (m *Contest) GetMultis() *Multis {
	if m != nil {
		return m.Multis
	}
	return nil
}

func (m *Contest) GetXchangeMultiPattern() string {
	if m != nil {
		return m.XchangeMultiPattern
	}
	return ""
}

func (m *Contest) GetCountPerBand() bool {
	if m != nil {
		return m.CountPerBand
	}
	return false
}

func (m *Contest) GetCabrilloQsoTemplate() string {
	if m != nil {
		return m.CabrilloQsoTemplate
	}
	return ""
}

func (m *Contest) GetCallHistoryFilename() string {
	if m != nil {
		return m.CallHistoryFilename
	}
	return ""
}

func (m *Contest) GetDefinitionYaml() string {
	if m != nil {
		return m.DefinitionYaml
	}
	return ""
}

func (m *Contest) GetExchangeValues() []string {
	if m != nil {
		return m.ExchangeValues
	}
	return nil
}

func (m *Contest) GetGenerateSerialExchange() bool {
	if m != nil {
		return m.GenerateSerialExchange
	}
	return false
}

func (m *Contest) GetCallHistoryFieldNames() []string {
	if m != nil {
		return m.CallHistoryFieldNames
	}
	return nil
}

func (m *Contest) GetQsosGoal() int32 {
	if m != nil {
		return m.QsosGoal
	}
	return 0
}

func (m *Contest) GetPointsGoal() int32 {
	if m != nil {
		return m.PointsGoal
	}
	return 0
}

func (m *Contest) GetMultisGoal() int32 {
	if m != nil {
		return m.MultisGoal
	}
	return 0
}

func (m *Contest) GetSprintOperation() bool {
	if m != nil {
		return m.SprintOperation
	}
	return false
}

func (m *Contest) GetStartTime() *timestamppb.Timestamp {
	if m != nil {
		return m.StartTime
	}
	return nil
}

func (m *Contest) GetGenerateReport() bool {
	if m != nil {
		return m.GenerateReport
	}
	return false
}

type Multis struct {
	Dxcc                 bool     `protobuf:"varint,1,opt,name=dxcc,proto3" json:"dxcc,omitempty"`
	Wpx                  bool     `protobuf:"varint,2,opt,name=wpx,proto3" json:"wpx,omitempty"`
	Xchange              bool     `protobuf:"varint,3,opt,name=xchange,proto3" json:"xchange,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Multis) Reset()         { *m = Multis{} }
func (m *Multis) String() string { return proto.CompactTextString(m) }
func (*Multis) ProtoMessage()    {}
func (*Multis) Descriptor() ([]byte, []int) {
	return fileDescriptor_a153da538f858886, []int{5}
}

func (m *Multis) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Multis.Unmarshal(m, b)
}
func (m *Multis) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Multis.Marshal(b, m, deterministic)
}
func (m *Multis) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Multis.Merge(m, src)
}
func (m *Multis) XXX_Size() int {
	return xxx_messageInfo_Multis.Size(m)
}
func (m *Multis) XXX_DiscardUnknown() {
	xxx_messageInfo_Multis.DiscardUnknown(m)
}

var xxx_messageInfo_Multis proto.InternalMessageInfo

func (m *Multis) GetDxcc() bool {
	if m != nil {
		return m.Dxcc
	}
	return false
}

func (m *Multis) GetWpx() bool {
	if m != nil {
		return m.Wpx
	}
	return false
}

func (m *Multis) GetXchange() bool {
	if m != nil {
		return m.Xchange
	}
	return false
}

type Keyer struct {
	Wpm                  int32    `protobuf:"varint,1,opt,name=wpm,proto3" json:"wpm,omitempty"`
	SpMacros             []string `protobuf:"bytes,2,rep,name=sp_macros,json=spMacros,proto3" json:"sp_macros,omitempty"`
	RunMacros            []string `protobuf:"bytes,3,rep,name=run_macros,json=runMacros,proto3" json:"run_macros,omitempty"`
	SpLabels             []string `protobuf:"bytes,4,rep,name=sp_labels,json=spLabels,proto3" json:"sp_labels,omitempty"`
	RunLabels            []string `protobuf:"bytes,5,rep,name=run_labels,json=runLabels,proto3" json:"run_labels,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Keyer) Reset()         { *m = Keyer{} }
func (m *Keyer) String() string { return proto.CompactTextString(m) }
func (*Keyer) ProtoMessage()    {}
func (*Keyer) Descriptor() ([]byte, []int) {
	return fileDescriptor_a153da538f858886, []int{6}
}

func (m *Keyer) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Keyer.Unmarshal(m, b)
}
func (m *Keyer) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Keyer.Marshal(b, m, deterministic)
}
func (m *Keyer) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Keyer.Merge(m, src)
}
func (m *Keyer) XXX_Size() int {
	return xxx_messageInfo_Keyer.Size(m)
}
func (m *Keyer) XXX_DiscardUnknown() {
	xxx_messageInfo_Keyer.DiscardUnknown(m)
}

var xxx_messageInfo_Keyer proto.InternalMessageInfo

func (m *Keyer) GetWpm() int32 {
	if m != nil {
		return m.Wpm
	}
	return 0
}

func (m *Keyer) GetSpMacros() []string {
	if m != nil {
		return m.SpMacros
	}
	return nil
}

func (m *Keyer) GetRunMacros() []string {
	if m != nil {
		return m.RunMacros
	}
	return nil
}

func (m *Keyer) GetSpLabels() []string {
	if m != nil {
		return m.SpLabels
	}
	return nil
}

func (m *Keyer) GetRunLabels() []string {
	if m != nil {
		return m.RunLabels
	}
	return nil
}

func init() {
	proto.RegisterType((*FileInfo)(nil), "pb.FileInfo")
	proto.RegisterType((*Entry)(nil), "pb.Entry")
	proto.RegisterType((*QSO)(nil), "pb.QSO")
	proto.RegisterType((*Station)(nil), "pb.Station")
	proto.RegisterType((*Contest)(nil), "pb.Contest")
	proto.RegisterType((*Multis)(nil), "pb.Multis")
	proto.RegisterType((*Keyer)(nil), "pb.Keyer")
}

func init() {
	proto.RegisterFile("log.proto", fileDescriptor_a153da538f858886)
}

var fileDescriptor_a153da538f858886 = []byte{
	// 1056 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x55, 0xdd, 0x6e, 0xdb, 0x36,
	0x14, 0x8e, 0x63, 0x3b, 0xb6, 0x8e, 0x13, 0xc7, 0xa1, 0x97, 0x46, 0x4d, 0x3a, 0x34, 0xf5, 0x36,
	0x34, 0x03, 0x06, 0x17, 0xcb, 0x80, 0xfd, 0x5d, 0xb6, 0x68, 0x97, 0x75, 0x4b, 0x93, 0x28, 0x41,
	0xb1, 0x61, 0x17, 0x04, 0x2d, 0xd3, 0x8e, 0x30, 0x4a, 0x54, 0x48, 0xba, 0x8d, 0x1e, 0x63, 0x0f,
	0xb0, 0x9b, 0x3d, 0xc7, 0x1e, 0x6e, 0xe0, 0x21, 0x29, 0xd7, 0x1e, 0xd0, 0x3b, 0xf2, 0xfb, 0xe1,
	0xcf, 0xe1, 0x39, 0x87, 0x10, 0x09, 0x39, 0x1f, 0x97, 0x4a, 0x1a, 0x49, 0x36, 0xcb, 0xc9, 0xe1,
	0xe3, 0xb9, 0x94, 0x73, 0xc1, 0x9f, 0x21, 0x32, 0x59, 0xcc, 0x9e, 0x99, 0x2c, 0xe7, 0xda, 0xb0,
	0xbc, 0x74, 0xa2, 0xd1, 0xd7, 0xd0, 0x7d, 0x95, 0x09, 0xfe, 0x73, 0x31, 0x93, 0xe4, 0x0b, 0xe8,
	0xcf, 0xa4, 0xca, 0x99, 0xa1, 0xef, 0xb8, 0xd2, 0x99, 0x2c, 0xe2, 0xc6, 0x71, 0xe3, 0xa4, 0x9d,
	0xec, 0x38, 0xf4, 0xad, 0x03, 0x47, 0xff, 0x34, 0xa0, 0xfd, 0xb2, 0x30, 0xaa, 0x22, 0x47, 0xd0,
	0xbc, 0xd3, 0x12, 0x55, 0xbd, 0xd3, 0xce, 0xb8, 0x9c, 0x8c, 0xaf, 0xae, 0x2f, 0xce, 0x36, 0x12,
	0x8b, 0x92, 0xa7, 0xd0, 0xd1, 0x86, 0x19, 0xbb, 0xcc, 0x26, 0x0a, 0x7a, 0x56, 0x70, 0xed, 0xa0,
	0xb3, 0x8d, 0x24, 0xb0, 0x56, 0x98, 0xca, 0xc2, 0x70, 0x6d, 0xe2, 0xe6, 0x52, 0xf8, 0xc2, 0x41,
	0x56, 0xe8, 0x59, 0xf2, 0x04, 0xda, 0x7f, 0xf2, 0x8a, 0xab, 0xb8, 0x85, 0xb2, 0xc8, 0xca, 0x7e,
	0xb1, 0xc0, 0xd9, 0x46, 0xe2, 0x98, 0xe7, 0x1d, 0x68, 0x73, 0x7b, 0xb4, 0xd1, 0xbf, 0x4d, 0x68,
	0x5e, 0x5d, 0x5f, 0x90, 0x43, 0xe8, 0xa6, 0x4c, 0x08, 0x9d, 0xcd, 0xdd, 0x6d, 0xa2, 0xa4, 0x9e,
	0x93, 0x47, 0x10, 0xd5, 0xe1, 0xc0, 0x33, 0x36, 0x93, 0x25, 0x40, 0x08, 0xb4, 0x26, 0xac, 0x98,
	0xe2, 0x99, 0xa2, 0x04, 0xc7, 0x16, 0xcb, 0xe5, 0x94, 0xe3, 0x01, 0xa2, 0x04, 0xc7, 0xe4, 0x08,
	0xa2, 0xbc, 0xa2, 0x8a, 0x97, 0x52, 0x99, 0xb8, 0xed, 0xb6, 0xc8, 0xab, 0x04, 0xe7, 0x9e, 0x2c,
	0x16, 0xf9, 0x84, 0xab, 0x78, 0x0b, 0xa3, 0xd9, 0xcd, 0xab, 0x37, 0x38, 0x27, 0x4f, 0x60, 0xdb,
	0xdc, 0xf2, 0x4c, 0x05, 0x73, 0x07, 0xcd, 0x3d, 0xc4, 0xbc, 0xbf, 0x96, 0xf8, 0x25, 0xba, 0xb8,
	0x84, 0x93, 0xf8, 0x55, 0x3e, 0x83, 0x1d, 0x21, 0xe7, 0x74, 0x79, 0x93, 0x08, 0x6f, 0xb2, 0x2d,
	0xe4, 0xfc, 0xa6, 0xbe, 0xcc, 0xa7, 0x00, 0x79, 0x45, 0xef, 0xd3, 0x5b, 0x56, 0xcc, 0x79, 0x0c,
	0xb8, 0x51, 0x94, 0x57, 0xbf, 0x39, 0xc0, 0xae, 0xe1, 0xb6, 0x09, 0x8a, 0x1e, 0x2a, 0xdc, 0xde,
	0x41, 0xf4, 0x08, 0xa2, 0x99, 0xe2, 0x77, 0x0b, 0x5e, 0xa4, 0x55, 0xbc, 0x7d, 0xdc, 0x38, 0x69,
	0x24, 0x4b, 0x80, 0x3c, 0x86, 0x5e, 0x5e, 0x51, 0x1e, 0x16, 0xe8, 0x1f, 0x37, 0x4f, 0xa2, 0x04,
	0xf2, 0xea, 0xa5, 0x47, 0x6c, 0x76, 0xb9, 0x3d, 0x6a, 0xcd, 0x2e, 0x6a, 0xdc, 0xce, 0x41, 0xf6,
	0xba, 0xd5, 0xdd, 0x19, 0xf4, 0x47, 0x7f, 0x40, 0xc7, 0x67, 0xca, 0x47, 0x5f, 0xf0, 0x10, 0xba,
	0xb2, 0xe4, 0x8a, 0x19, 0xa9, 0xf0, 0x01, 0xa3, 0xa4, 0x9e, 0x93, 0x18, 0x3a, 0x42, 0xa6, 0x48,
	0xb9, 0x27, 0x0c, 0xd3, 0xd1, 0xdf, 0x11, 0x74, 0x7c, 0x7a, 0xd9, 0x17, 0x2d, 0x58, 0xce, 0xfd,
	0xca, 0x38, 0x26, 0x5f, 0x01, 0xe1, 0x85, 0xe1, 0x8a, 0xae, 0x84, 0xde, 0xae, 0xdf, 0x4d, 0x06,
	0xc8, 0xdc, 0x7c, 0x10, 0xff, 0x31, 0x0c, 0x3f, 0x54, 0x87, 0xcb, 0x35, 0x51, 0xbe, 0xb7, 0x94,
	0x87, 0x30, 0x9e, 0xc2, 0xbe, 0x0d, 0x5a, 0xa6, 0xf8, 0x9a, 0xa3, 0x85, 0x8e, 0xa1, 0x27, 0x57,
	0x3c, 0x27, 0x30, 0x60, 0x42, 0xc8, 0xf7, 0x34, 0x5f, 0x08, 0x93, 0x51, 0xcc, 0xcb, 0x36, 0xca,
	0xfb, 0x88, 0x9f, 0x5b, 0xf8, 0xb9, 0xcd, 0xd0, 0x35, 0x25, 0x66, 0xeb, 0xd6, 0xba, 0xf2, 0xdc,
	0xe6, 0xed, 0x18, 0x86, 0x9a, 0xe5, 0x9c, 0xa6, 0x72, 0x61, 0x2b, 0x86, 0x96, 0x32, 0x2b, 0x8c,
	0xc6, 0x24, 0x6c, 0x27, 0x7b, 0x96, 0x7a, 0xe1, 0x98, 0x4b, 0x24, 0xec, 0xb9, 0xbd, 0xbe, 0x30,
	0x59, 0xc1, 0x0b, 0x13, 0x1c, 0x2e, 0x27, 0x87, 0xce, 0xe1, 0x39, 0xef, 0xf9, 0x16, 0x0e, 0x74,
	0xc9, 0xd3, 0x6c, 0x96, 0xa5, 0xeb, 0xfb, 0x44, 0xe8, 0xda, 0x0f, 0xf4, 0xea, 0x5e, 0x3f, 0xc2,
	0xc3, 0xff, 0xfb, 0x14, 0x9f, 0x65, 0xf7, 0x5c, 0xc7, 0x80, 0x69, 0x73, 0xb0, 0xee, 0xf4, 0xb4,
	0x2d, 0x19, 0x69, 0x6e, 0xb9, 0x0a, 0x1b, 0xf5, 0x5c, 0xc9, 0x20, 0xe6, 0x97, 0x1f, 0xc1, 0x16,
	0x86, 0x47, 0x63, 0x1a, 0xf7, 0x4e, 0xc1, 0x76, 0x12, 0x8c, 0x8c, 0x4e, 0x3c, 0x63, 0xaf, 0xeb,
	0x1f, 0xc6, 0x87, 0xb2, 0x64, 0xc6, 0x70, 0x55, 0xc4, 0x3b, 0x98, 0x29, 0x43, 0x4f, 0xa2, 0xeb,
	0xd2, 0x51, 0xe4, 0x73, 0xe8, 0xe3, 0x69, 0x69, 0xc9, 0x95, 0x7b, 0xa4, 0x3e, 0x86, 0x7e, 0x1b,
	0xd1, 0x4b, 0xae, 0xf0, 0x89, 0x4e, 0x61, 0x3f, 0x65, 0x13, 0x95, 0x09, 0x21, 0xe9, 0x9d, 0x96,
	0xd4, 0xf0, 0xbc, 0x14, 0xcc, 0xd8, 0x7a, 0xc0, 0x95, 0x03, 0x79, 0xa5, 0xe5, 0x8d, 0xa7, 0x9c,
	0x47, 0x08, 0x7a, 0x9b, 0x69, 0x23, 0x55, 0x45, 0x67, 0x99, 0xe0, 0x98, 0xb7, 0x83, 0xe0, 0x11,
	0xe2, 0xcc, 0x71, 0xaf, 0x3c, 0x45, 0x9e, 0xc2, 0xee, 0x94, 0xcf, 0xb2, 0x22, 0xb3, 0x65, 0x44,
	0x2b, 0x96, 0x8b, 0x98, 0xa0, 0xba, 0xbf, 0x84, 0x7f, 0x67, 0xb9, 0xb0, 0xc2, 0x50, 0x93, 0xf4,
	0x1d, 0x13, 0x0b, 0xae, 0xe3, 0x21, 0xc6, 0xb8, 0x1f, 0xe0, 0xb7, 0x88, 0x92, 0xef, 0x21, 0x9e,
	0xf3, 0xc2, 0xd6, 0x17, 0xa7, 0x9a, 0xab, 0x8c, 0x89, 0x65, 0x31, 0x7f, 0x82, 0x37, 0x7d, 0x10,
	0xf8, 0x6b, 0xa4, 0xeb, 0xe2, 0xff, 0x0e, 0xe2, 0xb5, 0xf3, 0x73, 0x31, 0xa5, 0xf6, 0x98, 0x3a,
	0xde, 0xc7, 0xbd, 0xf6, 0x57, 0xae, 0xc0, 0xc5, 0xf4, 0x8d, 0x25, 0x6d, 0x03, 0xbd, 0xd3, 0x52,
	0xd3, 0xb9, 0x64, 0x22, 0x7e, 0xe0, 0x1a, 0xa8, 0x05, 0x7e, 0x92, 0x4c, 0xd8, 0x9e, 0xe3, 0x1e,
	0xd9, 0xd1, 0x07, 0x48, 0x83, 0x83, 0x82, 0xc0, 0x3d, 0xa7, 0x13, 0xc4, 0x4e, 0xe0, 0x20, 0x14,
	0x7c, 0x09, 0x03, 0x5d, 0xaa, 0xac, 0x30, 0xd4, 0xf5, 0x0d, 0xfb, 0x5b, 0x3d, 0xc4, 0x9b, 0xec,
	0x3a, 0xfc, 0x22, 0xc0, 0xe4, 0x07, 0x00, 0x6d, 0x98, 0x32, 0xd8, 0x69, 0xe3, 0x43, 0x4c, 0x9c,
	0xc3, 0xb1, 0xfb, 0x5f, 0xc7, 0xe1, 0x7f, 0x1d, 0xd7, 0x2d, 0x37, 0x89, 0x50, 0x6d, 0xe7, 0x36,
	0xc0, 0x75, 0xdc, 0x7c, 0xaf, 0x3f, 0x72, 0x35, 0x19, 0x60, 0xd7, 0xee, 0x5f, 0xb7, 0xba, 0x7b,
	0x03, 0x32, 0x3a, 0x83, 0x2d, 0x97, 0x8c, 0xb6, 0x3b, 0x4d, 0xef, 0xd3, 0x14, 0xbb, 0x53, 0x37,
	0xc1, 0x31, 0x19, 0x40, 0xf3, 0x7d, 0x79, 0xef, 0xdb, 0x91, 0x1d, 0xda, 0x4e, 0xb7, 0xda, 0x75,
	0xc2, 0x74, 0xf4, 0x57, 0x03, 0xda, 0xf8, 0x43, 0x3a, 0x57, 0xee, 0x3f, 0x74, 0x3b, 0xb4, 0x91,
	0xd5, 0x25, 0xcd, 0x59, 0xaa, 0xa4, 0x8e, 0x37, 0xf1, 0x0d, 0xba, 0xba, 0x3c, 0xc7, 0xb9, 0xfd,
	0x2f, 0xd4, 0xa2, 0x08, 0x6c, 0x13, 0xd9, 0x48, 0x2d, 0x0a, 0x4f, 0x3b, 0xaf, 0x60, 0x13, 0x2e,
	0x74, 0xdc, 0x0a, 0xde, 0x5f, 0x71, 0x1e, 0xbc, 0x9e, 0x6d, 0xd7, 0x5e, 0x47, 0x4f, 0xb6, 0x30,
	0x56, 0xdf, 0xfc, 0x17, 0x00, 0x00, 0xff, 0xff, 0x20, 0x33, 0x08, 0xe4, 0xaa, 0x08, 0x00, 0x00,
}
