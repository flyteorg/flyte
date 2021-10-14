{{range .}}{{if not .ID}}{{else}}
<a href="{{.HTMLURL}}">
    <img src="{{.AvatarURL}}" width="64" height="64" alt="{{.ID}}">
</a>{{end}}
{{end}}