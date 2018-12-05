<html>
    <head>Cache.config</head>
    <body>
        <form action="/exec" method="post">
            url: <input type="text" name="url" value="{{.URLString}}" /> 
            <input type="submit" name="submit" />
        </form>
        Matched Rules <br />
        <table border="1">
            <tr>
                <td>Line</td>
                <td>Rule</td>
            </tr>
        {{with .MatchedRules}}
            {{range .}}
            <tr>
                <td>{{.Line}}</td>
                <td>{{.Rule}}</td>
            </tr>
            {{end}}
        {{end}}
        </table>
        <br />
        Error Rules <br />
        <table border="1">
            <tr>
                <td>Line</td>
                <td>Rule</td>
            </tr>
        {{with .ErrorRules}}
            {{range .}}
            <tr>
                <td>{{.Line}}</td>
                <td>{{.Rule}}</td>
            </tr>
            {{end}}
        {{end}}
        </table>
    </body>
</html>