{{range .}}
{{if not .}}
{{else}}
<a href="{{.HTMLURL}}">
    <img src="{{.AvatarURL}}" width="64" height="64" alt="{{.ID}}" style="border-radius: 50%;">
</a>
{{end}}
{{end}}
