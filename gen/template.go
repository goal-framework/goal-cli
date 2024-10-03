package gen

import (
	"fmt"
	"github.com/goal-web/supports/logs"
	"log"
	"os"
	"strings"
	"text/template"
)

var defaultTemplate = []byte("{{- define \"model\" -}}\npackage {{ .Package }}\n  \nimport (\n  \"github.com/goal-web/contracts\"\n  \"github.com/goal-web/database/table\"\n  \"github.com/goal-web/supports/utils\"\n  {{- if .Model.Authenticatable }}\n  \"fmt\"\n  {{- end }}\n\n  {{- range .Imports }}\n{{ .Alias }} \"{{ .Pkg }}\"\n{{- end }}\n)\n\n{{- $modelName := .Model.Name }}\n{{- $tableName := .Model.TableName }}\n{{- $primaryKey := .Model.PrimaryKey }}\n\nfunc New{{ $modelName }}(fields contracts.Fields) *{{ $modelName }} {\n  var model {{ $modelName }}\n  model.Set(fields)\n  return &model\n}\n\nfunc {{ .Model.RawName }}Query() *table.Table[{{ $modelName }}] {\n  return table.NewQuery(\"{{ $tableName }}\", New{{ $modelName }}).SetPrimaryKey(\"{{ $primaryKey }}\")\n}\n\n{{ toComments .Model.Name .Model.Comments }}\ntype {{ $modelName }} struct {\n  {{- range .Fields }}\n  {{ .Comment }}\n  {{ .Name }} {{ goType . }} `{{ toTags . }}`\n  {{- end }}\n  \n  _update contracts.Fields\n}\n\nfunc (model *{{ $modelName }}) Exists() bool {\n  return {{ .Model.RawName }}Query().Where(\"{{ $primaryKey }}\", model.GetPrimaryKey()).Count() > 0\n}\n\nfunc (model *{{ $modelName }}) Save() contracts.Exception {\n  if model._update == nil {\n    return nil\n  }\n  if {{ $modelName }}Saving != nil {\n    if err := {{ $modelName }}Saving(model); err != nil {\n      return err\n    }\n  } \n  _, err := {{ .Model.RawName }}Query().Where(\"{{ $primaryKey }}\", model.GetPrimaryKey()).UpdateE(model._update)\n  if err == nil {\n    model._update = nil\n    if {{ $modelName }}Saved != nil {\n      {{ $modelName }}Saved(model)\n    }\n  }\n  \n  return err\n}\n\nfunc (model *{{ $modelName }}) Set(fields contracts.Fields) {\n  for key, value := range fields {\n  {{- range .Fields }}\n    if key == \"{{ .JSONName }}\" {\n      switch v := value.(type) {\n      case {{ goType . }}:\n        model.Set{{ .Name }}(v)\n      case func() {{ goType . }}:\n        model.Set{{ .Name }}(v())\n      }\n    }\n  {{- end }}\n  }\n}\n\nfunc (model *{{ $modelName }}) Only(key ...string) contracts.Fields {\n  var fields = make(contracts.Fields)\n  for _, k := range key {\n  {{- range .Fields }}\n    if k == \"{{ .JSONName }}\" {\n      fields[k] = model.Get{{ .Name }}()\n      continue\n    }\n  {{- end }}\n  \n    if {{ $modelName }}Appends[k] != nil {\n     fields[k] = {{ $modelName }}Appends[k](model)\n    }\n  }\n  return fields\n}\n\nfunc (model *{{ $modelName }}) Except(keys ...string) contracts.Fields {\n  var excepts = map[string]struct{}{}\n  for _, k := range keys {\n    excepts[k] = struct{}{}\n  }\n  var fields = make(contracts.Fields)\n  for key, value := range model.ToFields() {\n    if _, ok := excepts[key]; ok {\n      continue\n    }\n    fields[key] = value\n  }\n  return fields\n}\n\n\nvar {{ $modelName }}Appends = map[string]func(model *{{ $modelName }}) any{}\n  \nfunc (model *{{ $modelName }}) ToFields() contracts.Fields {\n  fields := contracts.Fields{\n  {{- range .Fields }}\n    \"{{ .JSONName }}\": model.Get{{ .Name }}(),\n  {{- end }}\n  }\n\n  for key, f := range {{ $modelName }}Appends {\n    fields[key] = f(model)\n  }\n\n\n  return fields\n}\n\nfunc (model *{{ $modelName }}) Update(fields contracts.Fields) contracts.Exception {\n\n  if {{ $modelName }}Updating != nil {\n    if err := {{ $modelName }}Updating(model, fields); err != nil {\n      return err\n    }\n  }\n\n  if model._update != nil {\n    utils.MergeFields(model._update, fields)\n  }\n\n\n  _, err := {{ .Model.RawName }}Query().Where(\"{{ $primaryKey }}\", model.GetPrimaryKey()).UpdateE(fields)\n  \n  if err == nil {\n    model.Set(fields)\n    model._update = nil\n    if {{ $modelName }}Updated != nil {\n      {{ $modelName }}Updated(model, fields)\n    }\n  }\n  \n  return err\n}\n\nfunc (model *{{ $modelName }}) Refresh() contracts.Exception {\n  fields, err := table.ArrayQuery(\"{{ $tableName }}\").Where(\"{{ $primaryKey }}\", model.GetPrimaryKey()).FirstE()\n  if err != nil {\n    return err\n  }\n\n  model.Set(*fields)\n  return nil\n}\n\nfunc (model *{{ $modelName }}) Delete() contracts.Exception {\n  \n  if {{ $modelName }}Deleting != nil {\n    if err := {{ $modelName }}Deleting(model); err != nil {\n      return err\n    }\n  }\n\n  _, err := {{ .Model.RawName }}Query().Where(\"{{ $primaryKey }}\", model.GetPrimaryKey()).DeleteE()\n  if err == nil && {{ $modelName }}Deleted != nil {\n    {{ $modelName }}Deleted(model)\n  }\n  \n  return err\n}\n\nvar (\n  {{- range .Fields }}\n  {{ $modelName }}{{ .Name }}Getter func(model *{{ $modelName }}, raw {{ goType . }}) {{ goType . }}\n  {{ $modelName }}{{ .Name }}Setter func(model *{{ $modelName }}, raw {{ goType . }}) {{ goType . }}\n  {{- end }}\n  {{ $modelName }}Saving   func(model *{{ $modelName }}) contracts.Exception\n  {{ $modelName }}Saved    func(model *{{ $modelName }})\n  {{ $modelName }}Updating func(model *{{ $modelName }}, fields contracts.Fields) contracts.Exception\n  {{ $modelName }}Updated  func(model *{{ $modelName }}, fields contracts.Fields)\n  {{ $modelName }}Deleting func(model *{{ $modelName }}) contracts.Exception\n  {{ $modelName }}Deleted  func(model *{{ $modelName }})\n  {{ $modelName }}PrimaryKeyGetter func(model *{{ $modelName }}) any\n)\n  \n  \nfunc (model *{{ $modelName }}) GetPrimaryKey() any {\n  if {{ $modelName }}PrimaryKeyGetter != nil {\n    return {{ $modelName }}PrimaryKeyGetter(model)\n  }\n  \n  return model.{{ toCamelCase $primaryKey }}\n}\n\n{{- if .Model.Authenticatable }}\nfunc (model *{{ $modelName }}) GetAuthenticatableKey() string {\n  return fmt.Sprintf(\"%v\", model.GetPrimaryKey())\n}\n\nfunc {{ .Model.RawName }}AuthProvider(identify string) contracts.Authenticatable {\n  return {{ .Model.RawName }}Query().Find(identify)\n}\n\n{{- end }}\n\n\n{{- range .Fields }}\n\nfunc (model *{{ $modelName }}) Get{{ .Name }}() {{ goType . }} {\n  if {{ $modelName }}{{ .Name }}Getter != nil {\n    return {{ $modelName }}{{ .Name }}Getter(model, model.{{ .Name }})\n  }\n  return model.{{ .Name }}\n}\n\nfunc (model *{{ $modelName }}) Set{{ .Name }}(value {{ goType . }}) {\n  if {{ $modelName }}{{ .Name }}Setter != nil {\n    value = {{ $modelName }}{{ .Name }}Setter(model, value)\n  }\n\n  if model._update == nil {\n    model._update = contracts.Fields{\"{{ .JSONName }}\": value}\n  } else {\n    model._update[\"{{ .JSONName }}\"] = value\n  }\n  model.{{ .Name }} = value\n}\n\n{{- end }}\n\n{{ end }}\n\n\n{{- define \"data\" -}}\npackage {{ .Package }}\n  \nimport (\n{{- range .Imports }}\n{{ .Alias }} \"{{ .Pkg }}\"\n{{- end }}\n)\n\ntype {{ .Model.Name }} struct {\n  {{- range .Fields }}\n  {{ .Name }} {{ goType . }} `{{ toTags . }}`\n  {{- end }}\n}\n\n{{ end }}\n\n{{- define \"request\" -}}\npackage {{ .Package }}\n  \nimport (\n  {{- range .Imports }}\n  {{ .Alias }} \"{{ .Pkg }}\"\n  {{- end }}\n  \"github.com/goal-web/contracts\"\n)\n\ntype {{ .Model.Name }} struct {\n  {{- range .Fields }}\n  {{ .Name }} {{ goType . }} `{{ toTags . }}`\n  {{- end }}\n}\n\nfunc (model *{{ .Model.Name }}) ToFields() contracts.Fields {\n  return contracts.Fields{\n  {{- range .Fields }}\n    \"{{ .JSONName }}\": model.{{ .Name }},\n  {{- end }}\n  }\n}\n\n{{ end }}\n\n{{- define \"result\" -}}\npackage {{ .Package }}\n    \nimport (\n  {{- range .Imports }}\n  {{ .Alias }} \"{{ .Pkg }}\"\n  {{- end }}\n)\n\ntype {{ .Model.Name }} struct {\n  {{- range .Fields }}\n  {{ .Name }} {{ goType . }} `{{ toTags . }}`\n  {{- end }}\n}\n\n{{ end }}\n\n{{- define \"enum\" -}}\npackage {{ .Package }}\n\n{{- $enumName := .Name }}\ntype {{ .Name }} int\nconst (\n  {{- range .Values }}\n  {{- $FieldName := sprintf \"%s%s\" $enumName .Name }}\n\n  {{ toComments $FieldName .Comments }}\n  {{ $enumName }}{{ .Name }} {{ $enumName }} = {{ .Value }}\n  {{- end }}\n  {{ $enumName }}Unknown {{ $enumName }} = -1000\n      \n)\n  \n  \nfunc (item {{ $enumName }}) String() string {\n    switch item {\n      {{- range .Values }}\n        case {{ $enumName }}{{ .Name }}:\n          return \"{{ .Name }}\"\n      {{- end }}\n        default:\n          return \"Unknown\"\n  }\n}\n\nfunc (item {{ $enumName }}) Message() string {\n    switch item {\n      {{- range .Values }}\n        case {{ $enumName }}{{ .Name }}:\n          return \"{{ .Message }}\"\n      {{- end }}\n        default:\n          return \"Unknown\"\n  }\n}\n\nfunc Parse{{ $enumName }}FromString(msg string) {{ $enumName }} {\n    switch msg {\n    {{- range .Values }}\n        case \"{{ .Name }}\":\n          return {{ $enumName }}{{ .Name }}\n    {{- end }}\n        default:\n          return {{ $enumName }}Unknown\n  }\n}\n\n\n{{ end }}\n\n\n\n{{- define \"service\" -}}\npackage {{ .Package }}\n\nimport (\n  {{- range .Imports }}\n  {{ .Alias }} \"{{ .Pkg }}\"\n  {{- end }}\n)\n\n{{- $serviceName := .Name }}\n\n{{- range .Methods }}\n\nvar {{ $serviceName }}{{ .Name }}Handler func (req *{{ .InputUsageName }}) (*{{ .OutputUsageName }}, error)\nfunc {{ $serviceName }}{{ .Name }}(req *{{ .InputUsageName }}) (*{{ .OutputUsageName }}, error) {\n  if {{ $serviceName }}{{ .Name }}Handler != nil {\n    return {{ $serviceName }}{{ .Name }}Handler(req)\n  }\n  return nil, nil\n}\n{{- end }}\n{{ end }}\n\n{{- define \"controller\" -}}\npackage {{ .Package }}\n\nimport (\n  \"github.com/goal-web/contracts\"\n  \"github.com/goal-web/validation\"\n  \"{{ .ResponsePath }}\"\n  svc \"{{ .ImportPath }}\"\n  {{- range .Imports }}\n  {{- if notContains .Pkg \"results\" }}\n  {{ .Alias }} \"{{ .Pkg }}\"\n  {{ end -}}\n  {{- end }}\n)\n\n{{- $serviceName := .Name }}\n{{- $prefix := .Prefix }}\nfunc {{ .Name }}Router(router contracts.HttpRouter) {\n  routeGroup := router.Group(\"{{ $prefix }}\"{{ toMiddlewares .Middlewares }})\n  {{- range .Methods }}\n  {{- $controllerMethod := sprintf \"%s%s\" $serviceName .Name  }}\n  {{- $path := .Path  }}\n  {{- $middlewares := .Middlewares }}\n    {{- range .Method }}\n    routeGroup.{{ . }}(\"{{ $path }}\", {{ $controllerMethod }}{{ toMiddlewares $middlewares }})\n    {{- end }}\n  {{- end }}\n}\n\n\n{{- $usageName := .UsageName }}\n\n{{- range .Methods }}\nfunc {{ $serviceName }}{{ .Name }}(request contracts.HttpRequest) any {\n    var req {{ .InputUsageName }}\n    \n    if err:= request.Parse(&req); err != nil {\n      return response.ParseReqErr(err)\n    }\n    \n    if err := validation.Struct(req); err != nil {\n      return response.InvalidReq(err)\n    }\n  \n    resp, err := {{ $usageName }}{{ .Name }}(&req)\n    if err != nil {\n      return response.BizErr(err)\n    }\n    \n    return response.Success(resp)\n}\n{{- end }}\n{{ end }}")

func GetTemplate(path string) *template.Template {
	// 读取模板文件
	tmplContent, err := os.ReadFile(path)
	if err != nil {
		logs.Default().WithField("path", path).Warn("模板文件不存在，将使用默认模板")
		tmplContent = defaultTemplate
	}

	// 初始化模板，并添加函数映射
	tmpl, err := template.New("codegen").Funcs(template.FuncMap{
		"goType":        GoType,
		"toCamelCase":   ToCamelCase,
		"toSnake":       ToSnakeCase,
		"toTags":        ToTags,
		"replaceSuffix": strings.ReplaceAll,
		"toComments":    ToComments,
		"sprintf":       fmt.Sprintf,
		"contains":      strings.Contains,
		"notContains":   NotContains,
		"toMiddlewares": ToMiddlewares,
		"getComment":    GetComment,
		"hasComment":    HasComment,
		"hasMsgComment": HasMsgComment,
	}).Parse(string(tmplContent))
	if err != nil {
		log.Fatal(err)
	}
	return tmpl
}
