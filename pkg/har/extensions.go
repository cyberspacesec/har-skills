package har

import (
	"encoding/json"
	"strings"
)

// CustomFields 存储 HAR 规范中允许的以 "_" 为前缀的自定义扩展字段。
// HAR 规范允许任何以 "_" 开头的字段名作为自定义扩展数据。
// 例如：Chrome 的 "_initiator", "_priority", "_resourceType" 等。
type CustomFields map[string]interface{}

// GetCustomField 获取自定义扩展字段的值。
// 如果字段不存在，返回 nil。
func (cf CustomFields) GetCustomField(name string) interface{} {
	if cf == nil {
		return nil
	}
	return cf[name]
}

// SetCustomField 设置自定义扩展字段的值。
// 字段名应以 "_" 开头（符合 HAR 规范），但不会强制检查。
func (cf CustomFields) SetCustomField(name string, value interface{}) {
	if cf == nil {
		return
	}
	cf[name] = value
}

// HasCustomField 检查是否存在指定的自定义扩展字段。
func (cf CustomFields) HasCustomField(name string) bool {
	if cf == nil {
		return false
	}
	_, ok := cf[name]
	return ok
}

// DeleteCustomField 删除指定的自定义扩展字段。
func (cf CustomFields) DeleteCustomField(name string) {
	if cf == nil {
		return
	}
	delete(cf, name)
}

// CustomFieldsKeys 返回所有自定义扩展字段的名称。
func (cf CustomFields) CustomFieldsKeys() []string {
	if cf == nil {
		return nil
	}
	keys := make([]string, 0, len(cf))
	for k := range cf {
		keys = append(keys, k)
	}
	return keys
}

// knownUnderscoreKeys tracks the _-prefixed JSON keys that are already
// handled as typed struct fields, so they are not duplicated in CustomFields.
var knownUnderscoreKeys = map[string]map[string]bool{
	"Response": {"_transferSize": true, "_error": true},
	"Timings":  {"_blocked_queueing": true, "_blocked_proxy": true},
	"Entries":  {"_initiator": true, "_priority": true, "_resourceType": true},
}

// extractCustomFields 从原始JSON数据中提取 "_" 前缀的自定义扩展字段，
// 排除已由结构体字段处理的已知扩展字段。
func extractCustomFields(data []byte, typeName string) CustomFields {
	if len(data) == 0 {
		return nil
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}

	known := knownUnderscoreKeys[typeName]
	cf := make(CustomFields)
	for key, value := range raw {
		if strings.HasPrefix(key, "_") {
			// Skip keys already handled by struct fields
			if known != nil && known[key] {
				continue
			}
			var v interface{}
			if err := json.Unmarshal(value, &v); err != nil {
				cf[key] = string(value)
			} else {
				cf[key] = v
			}
		}
	}

	if len(cf) == 0 {
		return nil
	}
	return cf
}

// mergeCustomFieldsIntoJSON 将自定义扩展字段合并到标准JSON输出中
func mergeCustomFieldsIntoJSON(stdData []byte, cf CustomFields) ([]byte, error) {
	if len(cf) == 0 {
		return stdData, nil
	}

	var result map[string]json.RawMessage
	if err := json.Unmarshal(stdData, &result); err != nil {
		return stdData, nil
	}

	for key, value := range cf {
		v, err := json.Marshal(value)
		if err != nil {
			continue
		}
		result[key] = v
	}

	return json.Marshal(result)
}

// --- Har ---

// GetCustomField 获取Har上的自定义扩展字段值
func (h *Har) GetCustomField(name string) interface{} {
	return h.CustomFields.GetCustomField(name)
}

// SetCustomField 设置Har上的自定义扩展字段值
func (h *Har) SetCustomField(name string, value interface{}) {
	if h.CustomFields == nil {
		h.CustomFields = make(CustomFields)
	}
	h.CustomFields.SetCustomField(name, value)
}

// UnmarshalJSON 自定义反序列化，提取 "_" 前缀的自定义扩展字段
func (h *Har) UnmarshalJSON(data []byte) error {
	type Alias Har
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(h),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	h.CustomFields = extractCustomFields(data, "Har")
	return nil
}

// MarshalJSON 自定义序列化，将自定义扩展字段合并到JSON输出中。
// 使用值接收器以确保 json.Marshal(h) 在 h 为值类型或指针时均能调用此方法。
func (h Har) MarshalJSON() ([]byte, error) {
	type Alias Har
	data, err := json.Marshal(Alias(h))
	if err != nil {
		return nil, err
	}
	return mergeCustomFieldsIntoJSON(data, h.CustomFields)
}

// --- Log ---

// GetCustomField 获取Log上的自定义扩展字段值
func (l *Log) GetCustomField(name string) interface{} {
	return l.CustomFields.GetCustomField(name)
}

// SetCustomField 设置Log上的自定义扩展字段值
func (l *Log) SetCustomField(name string, value interface{}) {
	if l.CustomFields == nil {
		l.CustomFields = make(CustomFields)
	}
	l.CustomFields.SetCustomField(name, value)
}

// UnmarshalJSON 自定义反序列化，提取 "_" 前缀的自定义扩展字段
func (l *Log) UnmarshalJSON(data []byte) error {
	type Alias Log
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(l),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	l.CustomFields = extractCustomFields(data, "Log")
	return nil
}

// MarshalJSON 自定义序列化，将自定义扩展字段合并到JSON输出中
func (l Log) MarshalJSON() ([]byte, error) {
	type Alias Log
	data, err := json.Marshal(Alias(l))
	if err != nil {
		return nil, err
	}
	return mergeCustomFieldsIntoJSON(data, l.CustomFields)
}

// --- Entries ---

// GetCustomField 获取Entries上的自定义扩展字段值
func (e *Entries) GetCustomField(name string) interface{} {
	return e.CustomFields.GetCustomField(name)
}

// SetCustomField 设置Entries上的自定义扩展字段值
func (e *Entries) SetCustomField(name string, value interface{}) {
	if e.CustomFields == nil {
		e.CustomFields = make(CustomFields)
	}
	e.CustomFields.SetCustomField(name, value)
}

// UnmarshalJSON 自定义反序列化，提取 "_" 前缀的自定义扩展字段
func (e *Entries) UnmarshalJSON(data []byte) error {
	type Alias Entries
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	e.CustomFields = extractCustomFields(data, "Entries")
	return nil
}

// MarshalJSON 自定义序列化，将自定义扩展字段合并到JSON输出中
func (e Entries) MarshalJSON() ([]byte, error) {
	type Alias Entries
	data, err := json.Marshal(Alias(e))
	if err != nil {
		return nil, err
	}
	return mergeCustomFieldsIntoJSON(data, e.CustomFields)
}

// --- Request ---

// GetCustomField 获取Request上的自定义扩展字段值
func (r *Request) GetCustomField(name string) interface{} {
	return r.CustomFields.GetCustomField(name)
}

// SetCustomField 设置Request上的自定义扩展字段值
func (r *Request) SetCustomField(name string, value interface{}) {
	if r.CustomFields == nil {
		r.CustomFields = make(CustomFields)
	}
	r.CustomFields.SetCustomField(name, value)
}

// UnmarshalJSON 自定义反序列化，提取 "_" 前缀的自定义扩展字段
func (r *Request) UnmarshalJSON(data []byte) error {
	type Alias Request
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	r.CustomFields = extractCustomFields(data, "Request")
	return nil
}

// MarshalJSON 自定义序列化，将自定义扩展字段合并到JSON输出中
func (r Request) MarshalJSON() ([]byte, error) {
	type Alias Request
	data, err := json.Marshal(Alias(r))
	if err != nil {
		return nil, err
	}
	return mergeCustomFieldsIntoJSON(data, r.CustomFields)
}

// --- Response ---

// GetCustomField 获取Response上的自定义扩展字段值
func (r *Response) GetCustomField(name string) interface{} {
	return r.CustomFields.GetCustomField(name)
}

// SetCustomField 设置Response上的自定义扩展字段值
func (r *Response) SetCustomField(name string, value interface{}) {
	if r.CustomFields == nil {
		r.CustomFields = make(CustomFields)
	}
	r.CustomFields.SetCustomField(name, value)
}

// UnmarshalJSON 自定义反序列化，提取 "_" 前缀的自定义扩展字段
// 注意：Response已有_transferSize和_error的struct字段，CustomFields只存储其他扩展字段
func (r *Response) UnmarshalJSON(data []byte) error {
	type Alias Response
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	r.CustomFields = extractCustomFields(data, "Response")
	return nil
}

// MarshalJSON 自定义序列化，将自定义扩展字段合并到JSON输出中
func (r Response) MarshalJSON() ([]byte, error) {
	type Alias Response
	data, err := json.Marshal(Alias(r))
	if err != nil {
		return nil, err
	}
	return mergeCustomFieldsIntoJSON(data, r.CustomFields)
}

// --- Content ---

// GetCustomField 获取Content上的自定义扩展字段值
func (c *Content) GetCustomField(name string) interface{} {
	return c.CustomFields.GetCustomField(name)
}

// SetCustomField 设置Content上的自定义扩展字段值
func (c *Content) SetCustomField(name string, value interface{}) {
	if c.CustomFields == nil {
		c.CustomFields = make(CustomFields)
	}
	c.CustomFields.SetCustomField(name, value)
}

// UnmarshalJSON 自定义反序列化，提取 "_" 前缀的自定义扩展字段
func (c *Content) UnmarshalJSON(data []byte) error {
	type Alias Content
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	c.CustomFields = extractCustomFields(data, "Content")
	return nil
}

// MarshalJSON 自定义序列化，将自定义扩展字段合并到JSON输出中
func (c Content) MarshalJSON() ([]byte, error) {
	type Alias Content
	data, err := json.Marshal(Alias(c))
	if err != nil {
		return nil, err
	}
	return mergeCustomFieldsIntoJSON(data, c.CustomFields)
}

// --- Cookie ---

// GetCustomField 获取Cookie上的自定义扩展字段值
func (c *Cookie) GetCustomField(name string) interface{} {
	return c.CustomFields.GetCustomField(name)
}

// SetCustomField 设置Cookie上的自定义扩展字段值
func (c *Cookie) SetCustomField(name string, value interface{}) {
	if c.CustomFields == nil {
		c.CustomFields = make(CustomFields)
	}
	c.CustomFields.SetCustomField(name, value)
}

// UnmarshalJSON 自定义反序列化，提取 "_" 前缀的自定义扩展字段
func (c *Cookie) UnmarshalJSON(data []byte) error {
	type Alias Cookie
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	c.CustomFields = extractCustomFields(data, "Cookie")
	return nil
}

// MarshalJSON 自定义序列化，将自定义扩展字段合并到JSON输出中
func (c Cookie) MarshalJSON() ([]byte, error) {
	type Alias Cookie
	data, err := json.Marshal(Alias(c))
	if err != nil {
		return nil, err
	}
	return mergeCustomFieldsIntoJSON(data, c.CustomFields)
}

// --- Pages ---

// GetCustomField 获取Pages上的自定义扩展字段值
func (p *Pages) GetCustomField(name string) interface{} {
	return p.CustomFields.GetCustomField(name)
}

// SetCustomField 设置Pages上的自定义扩展字段值
func (p *Pages) SetCustomField(name string, value interface{}) {
	if p.CustomFields == nil {
		p.CustomFields = make(CustomFields)
	}
	p.CustomFields.SetCustomField(name, value)
}

// UnmarshalJSON 自定义反序列化，提取 "_" 前缀的自定义扩展字段
func (p *Pages) UnmarshalJSON(data []byte) error {
	type Alias Pages
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	p.CustomFields = extractCustomFields(data, "Pages")
	return nil
}

// MarshalJSON 自定义序列化，将自定义扩展字段合并到JSON输出中
func (p Pages) MarshalJSON() ([]byte, error) {
	type Alias Pages
	data, err := json.Marshal(Alias(p))
	if err != nil {
		return nil, err
	}
	return mergeCustomFieldsIntoJSON(data, p.CustomFields)
}

// --- Timings ---

// GetCustomField 获取Timings上的自定义扩展字段值
func (t *Timings) GetCustomField(name string) interface{} {
	return t.CustomFields.GetCustomField(name)
}

// SetCustomField 设置Timings上的自定义扩展字段值
func (t *Timings) SetCustomField(name string, value interface{}) {
	if t.CustomFields == nil {
		t.CustomFields = make(CustomFields)
	}
	t.CustomFields.SetCustomField(name, value)
}

// UnmarshalJSON 自定义反序列化，提取 "_" 前缀的自定义扩展字段
// 注意：Timings已有_blocked_queueing和_blocked_proxy的struct字段
func (t *Timings) UnmarshalJSON(data []byte) error {
	type Alias Timings
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	t.CustomFields = extractCustomFields(data, "Timings")
	return nil
}

// MarshalJSON 自定义序列化，将自定义扩展字段合并到JSON输出中
func (t Timings) MarshalJSON() ([]byte, error) {
	type Alias Timings
	data, err := json.Marshal(Alias(t))
	if err != nil {
		return nil, err
	}
	return mergeCustomFieldsIntoJSON(data, t.CustomFields)
}

// --- Cache ---

// GetCustomField 获取Cache上的自定义扩展字段值
func (c *Cache) GetCustomField(name string) interface{} {
	return c.CustomFields.GetCustomField(name)
}

// SetCustomField 设置Cache上的自定义扩展字段值
func (c *Cache) SetCustomField(name string, value interface{}) {
	if c.CustomFields == nil {
		c.CustomFields = make(CustomFields)
	}
	c.CustomFields.SetCustomField(name, value)
}

// UnmarshalJSON 自定义反序列化，提取 "_" 前缀的自定义扩展字段
func (c *Cache) UnmarshalJSON(data []byte) error {
	type Alias Cache
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	c.CustomFields = extractCustomFields(data, "Cache")
	return nil
}

// MarshalJSON 自定义序列化，将自定义扩展字段合并到JSON输出中
func (c Cache) MarshalJSON() ([]byte, error) {
	type Alias Cache
	data, err := json.Marshal(Alias(c))
	if err != nil {
		return nil, err
	}
	return mergeCustomFieldsIntoJSON(data, c.CustomFields)
}
