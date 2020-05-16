package main

const docIndexTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Docs</title>
    <style>
        * {
            padding: 0;
            margin: 0;
        }
    </style>
    <link rel="stylesheet" href="//cdn.jsdelivr.net/npm/docsify/lib/themes/vue.css" title="vue">
</head>
<body>
<section id="cover-section" class="cover show"
         style="background: linear-gradient(to left bottom, hsl(10, 100%, 85%) 0%,hsl(66, 100%, 85%) 100%);">
    <div class="cover-main">
        <h1><a><span>Docs Server</span></a></h1>
        <p>
            {{range .Projects}}
                <a href="{{.Link}}">{{.Name}}</a>
            {{end}}
            <a style="display: none" target="_blank"></a>
        </p>
    </div>
    <div class="mask"></div>
    <div style="position: absolute;bottom: 10px;left: 50%;transform: translateX(-50%)">
        <ul>
            <li>Powered by <a href="https://github.com/TheWinds/mkdoc" target="_blank">mkdoc</a> ❤️ <a
                        href="https://github.com/docsifyjs/docsify" target="_blank">docsify</a></li>
        </ul>
    </div>
</section>
<script>
    let i = Math.floor(Math.random() * 190) + 10
    let style = 'background: linear-gradient(to left bottom, hsl(' + i + ', 100%, 85%) 0%,hsl(49, 100%, 85%) 100%)'
    document.getElementById('cover-section').style = style

</script>
</body>
</html>`
