{{range .}}
{{if not .ID}}
{{else}}
<a href="https://images.weserv.nl/?url={{.HTMLURL}}&mask=circle">
    <img src="{{.AvatarURL}}" width="64" height="64" alt="{{.ID}}" style="border-radius: 50px;">
</a>
{{end}}
{{end}}