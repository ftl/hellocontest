// Code generated by protoc-gen-go. DO NOT EDIT.
// source: log.proto

package pb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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
	Name                    string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	EnterTheirNumber        bool     `protobuf:"varint,2,opt,name=enter_their_number,json=enterTheirNumber,proto3" json:"enter_their_number,omitempty"`
	EnterTheirXchange       bool     `protobuf:"varint,3,opt,name=enter_their_xchange,json=enterTheirXchange,proto3" json:"enter_their_xchange,omitempty"`
	RequireTheirXchange     bool     `protobuf:"varint,4,opt,name=require_their_xchange,json=requireTheirXchange,proto3" json:"require_their_xchange,omitempty"`
	AllowMultiBand          bool     `protobuf:"varint,5,opt,name=allow_multi_band,json=allowMultiBand,proto3" json:"allow_multi_band,omitempty"`
	AllowMultiMode          bool     `protobuf:"varint,6,opt,name=allow_multi_mode,json=allowMultiMode,proto3" json:"allow_multi_mode,omitempty"`
	SameCountryPoints       int32    `protobuf:"varint,7,opt,name=same_country_points,json=sameCountryPoints,proto3" json:"same_country_points,omitempty"`
	SameContinentPoints     int32    `protobuf:"varint,8,opt,name=same_continent_points,json=sameContinentPoints,proto3" json:"same_continent_points,omitempty"`
	SpecificCountryPoints   int32    `protobuf:"varint,9,opt,name=specific_country_points,json=specificCountryPoints,proto3" json:"specific_country_points,omitempty"`
	SpecificCountryPrefixes []string `protobuf:"bytes,10,rep,name=specific_country_prefixes,json=specificCountryPrefixes,proto3" json:"specific_country_prefixes,omitempty"`
	OtherPoints             int32    `protobuf:"varint,11,opt,name=other_points,json=otherPoints,proto3" json:"other_points,omitempty"`
	Multis                  *Multis  `protobuf:"bytes,12,opt,name=multis,proto3" json:"multis,omitempty"`
	XchangeMultiPattern     string   `protobuf:"bytes,13,opt,name=xchange_multi_pattern,json=xchangeMultiPattern,proto3" json:"xchange_multi_pattern,omitempty"`
	CountPerBand            bool     `protobuf:"varint,14,opt,name=count_per_band,json=countPerBand,proto3" json:"count_per_band,omitempty"`
	CabrilloQsoTemplate     string   `protobuf:"bytes,15,opt,name=cabrillo_qso_template,json=cabrilloQsoTemplate,proto3" json:"cabrillo_qso_template,omitempty"`
	CallHistoryFilename     string   `protobuf:"bytes,16,opt,name=call_history_filename,json=callHistoryFilename,proto3" json:"call_history_filename,omitempty"`
	DefinitionYaml          string   `protobuf:"bytes,18,opt,name=definition_yaml,json=definitionYaml,proto3" json:"definition_yaml,omitempty"`
	ExchangeValues          []string `protobuf:"bytes,19,rep,name=exchange_values,json=exchangeValues,proto3" json:"exchange_values,omitempty"`
	GenerateSerialExchange  bool     `protobuf:"varint,20,opt,name=generate_serial_exchange,json=generateSerialExchange,proto3" json:"generate_serial_exchange,omitempty"`
	CallHistoryFieldNames   []string `protobuf:"bytes,21,rep,name=call_history_field_names,json=callHistoryFieldNames,proto3" json:"call_history_field_names,omitempty"`
	XXX_NoUnkeyedLiteral    struct{} `json:"-"`
	XXX_unrecognized        []byte   `json:"-"`
	XXX_sizecache           int32    `json:"-"`
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
	// 920 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x55, 0xdd, 0x6e, 0x1b, 0x45,
	0x14, 0x4e, 0xe2, 0xd8, 0xde, 0x3d, 0x4e, 0x5c, 0x67, 0x8c, 0xe9, 0xd2, 0x16, 0x91, 0x1a, 0x50,
	0x73, 0x81, 0x2c, 0x11, 0x24, 0x40, 0x5c, 0xb6, 0x6a, 0x65, 0x8a, 0xd2, 0x26, 0x9b, 0x50, 0x81,
	0xb8, 0x18, 0xad, 0xd7, 0xc7, 0xce, 0x8a, 0x99, 0x9d, 0xcd, 0xcc, 0xb8, 0xcd, 0xbe, 0x0e, 0x2f,
	0xc0, 0x0b, 0xf0, 0x70, 0x68, 0xce, 0xcc, 0xda, 0xb5, 0x91, 0xb8, 0x9b, 0xf9, 0xbe, 0xef, 0xfc,
	0xcc, 0xf9, 0xd9, 0x85, 0x58, 0xa8, 0xe5, 0xa4, 0xd2, 0xca, 0x2a, 0x76, 0x50, 0xcd, 0xc6, 0xdf,
	0x42, 0xf4, 0xaa, 0x10, 0xf8, 0x73, 0xb9, 0x50, 0xec, 0x6b, 0xe8, 0x2f, 0x94, 0x96, 0x99, 0xe5,
	0xef, 0x51, 0x9b, 0x42, 0x95, 0xc9, 0xfe, 0xe9, 0xfe, 0x59, 0x3b, 0x3d, 0xf6, 0xe8, 0x3b, 0x0f,
	0x8e, 0xff, 0xda, 0x87, 0xf6, 0xcb, 0xd2, 0xea, 0x9a, 0x3d, 0x86, 0xd6, 0x9d, 0x51, 0xa4, 0xea,
	0x9d, 0x77, 0x27, 0xd5, 0x6c, 0x72, 0x75, 0xfd, 0x76, 0xba, 0x97, 0x3a, 0x94, 0x3d, 0x83, 0xae,
	0xb1, 0x99, 0x75, 0x6e, 0x0e, 0x48, 0xd0, 0x73, 0x82, 0x6b, 0x0f, 0x4d, 0xf7, 0xd2, 0x86, 0x75,
	0xc2, 0x5c, 0x95, 0x16, 0x8d, 0x4d, 0x5a, 0x1b, 0xe1, 0x0b, 0x0f, 0x39, 0x61, 0x60, 0xd9, 0x53,
	0x68, 0xff, 0x89, 0x35, 0xea, 0xe4, 0x90, 0x64, 0xb1, 0x93, 0xfd, 0xe2, 0x80, 0xe9, 0x5e, 0xea,
	0x99, 0xe7, 0x5d, 0x68, 0xa3, 0x4b, 0x6d, 0xfc, 0x4f, 0x0b, 0x5a, 0x57, 0xd7, 0x6f, 0xd9, 0x23,
	0x88, 0xf2, 0x4c, 0x08, 0x53, 0x2c, 0xfd, 0x6b, 0xe2, 0x74, 0x7d, 0x67, 0x4f, 0x20, 0xb6, 0x85,
	0x44, 0x63, 0x33, 0x59, 0x51, 0x8e, 0xad, 0x74, 0x03, 0x30, 0x06, 0x87, 0xb3, 0xac, 0x9c, 0x53,
	0x4e, 0x71, 0x4a, 0x67, 0x87, 0x49, 0x35, 0x47, 0x4a, 0x20, 0x4e, 0xe9, 0xcc, 0x1e, 0x43, 0x2c,
	0x6b, 0xae, 0xb1, 0x52, 0xda, 0x26, 0x6d, 0x1f, 0x42, 0xd6, 0x29, 0xdd, 0x03, 0x59, 0xae, 0xe4,
	0x0c, 0x75, 0xd2, 0xa1, 0x6a, 0x46, 0xb2, 0x7e, 0x43, 0x77, 0xf6, 0x14, 0x8e, 0xec, 0x2d, 0x16,
	0xba, 0x31, 0xee, 0x92, 0x71, 0x8f, 0xb0, 0x60, 0xbf, 0x96, 0x04, 0x17, 0x11, 0xb9, 0xf0, 0x92,
	0xe0, 0xe5, 0x4b, 0x38, 0x16, 0x6a, 0xc9, 0x37, 0x2f, 0x89, 0xe9, 0x25, 0x47, 0x42, 0x2d, 0x6f,
	0xd6, 0x8f, 0xf9, 0x1c, 0x40, 0xd6, 0xfc, 0x3e, 0xbf, 0xcd, 0xca, 0x25, 0x26, 0x40, 0x81, 0x62,
	0x59, 0xff, 0xe6, 0x01, 0xe7, 0xc3, 0x87, 0x69, 0x14, 0x3d, 0x52, 0xf8, 0xd8, 0x8d, 0xe8, 0x09,
	0xc4, 0x0b, 0x8d, 0x77, 0x2b, 0x2c, 0xf3, 0x3a, 0x39, 0x3a, 0xdd, 0x3f, 0xdb, 0x4f, 0x37, 0x00,
	0xfb, 0x02, 0x7a, 0xb2, 0xe6, 0xd8, 0x38, 0xe8, 0x9f, 0xb6, 0xce, 0xe2, 0x14, 0x64, 0xfd, 0x32,
	0x20, 0x6e, 0xba, 0x7c, 0x8c, 0xb5, 0xe6, 0x01, 0x69, 0x7c, 0xe4, 0x46, 0xf6, 0xfa, 0x30, 0x3a,
	0x1e, 0xf4, 0xc7, 0x7f, 0x40, 0x37, 0x4c, 0xca, 0xff, 0x76, 0xf0, 0x11, 0x44, 0xaa, 0x42, 0x9d,
	0x59, 0xa5, 0xa9, 0x81, 0x71, 0xba, 0xbe, 0xb3, 0x04, 0xba, 0x42, 0xe5, 0x44, 0xf9, 0x16, 0x36,
	0xd7, 0xf1, 0xdf, 0x5d, 0xe8, 0x86, 0xf1, 0x72, 0x1d, 0x2d, 0x33, 0x89, 0xc1, 0x33, 0x9d, 0xd9,
	0x37, 0xc0, 0xb0, 0xb4, 0xa8, 0xf9, 0x56, 0xe9, 0x9d, 0xff, 0x28, 0x1d, 0x10, 0x73, 0xf3, 0x51,
	0xfd, 0x27, 0x30, 0xfc, 0x58, 0xdd, 0x3c, 0xae, 0x45, 0xf2, 0x93, 0x8d, 0xbc, 0x29, 0xe3, 0x39,
	0x8c, 0x5c, 0xd1, 0x0a, 0x8d, 0x3b, 0x16, 0x87, 0x64, 0x31, 0x0c, 0xe4, 0x96, 0xcd, 0x19, 0x0c,
	0x32, 0x21, 0xd4, 0x07, 0x2e, 0x57, 0xc2, 0x16, 0x9c, 0xe6, 0xb2, 0x4d, 0xf2, 0x3e, 0xe1, 0x17,
	0x0e, 0x7e, 0xee, 0x26, 0x74, 0x47, 0x49, 0xd3, 0xda, 0xd9, 0x55, 0x5e, 0xb8, 0xb9, 0x9d, 0xc0,
	0xd0, 0x64, 0x12, 0x79, 0xae, 0x56, 0x6e, 0x63, 0x78, 0xa5, 0x8a, 0xd2, 0x1a, 0x1a, 0xc2, 0x76,
	0x7a, 0xe2, 0xa8, 0x17, 0x9e, 0xb9, 0x24, 0xc2, 0xe5, 0x1d, 0xf4, 0xa5, 0x2d, 0x4a, 0x2c, 0x6d,
	0x63, 0xe1, 0x67, 0x72, 0xe8, 0x2d, 0x02, 0x17, 0x6c, 0xbe, 0x87, 0x87, 0xa6, 0xc2, 0xbc, 0x58,
	0x14, 0xf9, 0x6e, 0x9c, 0x98, 0xac, 0x46, 0x0d, 0xbd, 0x1d, 0xeb, 0x27, 0xf8, 0xec, 0xbf, 0x76,
	0x1a, 0x17, 0xc5, 0x3d, 0x9a, 0x04, 0x68, 0x6c, 0x1e, 0xee, 0x5a, 0x06, 0xda, 0xad, 0x8c, 0xb2,
	0xb7, 0xa8, 0x9b, 0x40, 0x3d, 0xbf, 0x32, 0x84, 0x05, 0xf7, 0x63, 0xe8, 0x50, 0x79, 0x0c, 0x8d,
	0x71, 0xef, 0x1c, 0xdc, 0x97, 0x84, 0x2a, 0x63, 0xd2, 0xc0, 0xb8, 0xe7, 0x86, 0xc6, 0x84, 0x52,
	0x56, 0x99, 0xb5, 0xa8, 0xcb, 0xe4, 0x98, 0x26, 0x65, 0x18, 0x48, 0xb2, 0xba, 0xf4, 0x14, 0xfb,
	0x0a, 0xfa, 0x94, 0x2d, 0xaf, 0x50, 0xfb, 0x26, 0xf5, 0xa9, 0xf4, 0x47, 0x84, 0x5e, 0xa2, 0xa6,
	0x16, 0x9d, 0xc3, 0x28, 0xcf, 0x66, 0xba, 0x10, 0x42, 0xf1, 0x3b, 0xa3, 0xb8, 0x45, 0x59, 0x89,
	0xcc, 0xba, 0x7d, 0x20, 0xcf, 0x0d, 0x79, 0x65, 0xd4, 0x4d, 0xa0, 0xbc, 0x8d, 0x10, 0xfc, 0xb6,
	0x30, 0x56, 0xe9, 0x9a, 0x2f, 0x0a, 0x81, 0x34, 0xb7, 0x83, 0xc6, 0x46, 0x88, 0xa9, 0xe7, 0x5e,
	0x05, 0x8a, 0x3d, 0x83, 0x07, 0x73, 0x5c, 0x14, 0x65, 0xe1, 0xd6, 0x88, 0xd7, 0x99, 0x14, 0x09,
	0x23, 0x75, 0x7f, 0x03, 0xff, 0x9e, 0x49, 0xe1, 0x84, 0xcd, 0x4e, 0xf2, 0xf7, 0x99, 0x58, 0xa1,
	0x49, 0x86, 0x54, 0xe3, 0x7e, 0x03, 0xbf, 0x23, 0x94, 0xfd, 0x08, 0xc9, 0x12, 0x4b, 0xb7, 0x5f,
	0xc8, 0x0d, 0xea, 0x22, 0x13, 0x9b, 0x65, 0xfe, 0x84, 0x5e, 0xfa, 0x69, 0xc3, 0x5f, 0x13, 0xbd,
	0x5e, 0xfe, 0x1f, 0x20, 0xd9, 0xc9, 0x1f, 0xc5, 0x9c, 0xbb, 0x34, 0x4d, 0x32, 0xa2, 0x58, 0xa3,
	0xad, 0x27, 0xa0, 0x98, 0xbf, 0x71, 0xe4, 0xeb, 0xc3, 0xe8, 0x64, 0xc0, 0xc6, 0x53, 0xe8, 0xf8,
	0xf6, 0xb8, 0x7d, 0x9d, 0xdf, 0xe7, 0x39, 0xed, 0x6b, 0x94, 0xd2, 0x99, 0x0d, 0xa0, 0xf5, 0xa1,
	0xba, 0x0f, 0x0b, 0xea, 0x8e, 0x6e, 0xf7, 0xb7, 0xf7, 0xb0, 0xb9, 0x8e, 0x7f, 0x85, 0x36, 0xfd,
	0x32, 0xbc, 0x91, 0x0c, 0x7f, 0x38, 0x77, 0x74, 0xdf, 0x6a, 0x53, 0x71, 0x99, 0xe5, 0x5a, 0x99,
	0xe4, 0x80, 0x92, 0x8a, 0x4c, 0x75, 0x41, 0x77, 0xf7, 0x01, 0xd5, 0xab, 0xb2, 0x61, 0x5b, 0xc4,
	0xc6, 0x7a, 0x55, 0x7a, 0x7a, 0xd6, 0xa1, 0x3f, 0xea, 0x77, 0xff, 0x06, 0x00, 0x00, 0xff, 0xff,
	0x30, 0x7b, 0x54, 0x9c, 0x5e, 0x07, 0x00, 0x00,
}
