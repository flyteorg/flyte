{{range .}}{{if not .ID}}{{else}}
<a href="{{.HTMLURL}}">
    <img src="{{.AvatarURL}}" width="50" height="50" alt="{{.ID}}">
</a>{{end}}
{{end}}