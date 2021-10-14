{{range .}}
{{if not .ID}}
{{else}}
<a href="{{.HTMLURL}}">
    <img src="https://images.weserv.nl/?url={{.AvatarURL}}&mask=circle" width="64" height="64" alt="{{.ID}}">
</a>
{{end}}
{{end}}