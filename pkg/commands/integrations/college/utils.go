package integrations

var CollegeIntegrationID = "44b61604-49a8-4b4b-a868-86276cfdba62"

var EmailTemplate = `
<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta http-equiv="X-UA-Compatible" content="IE=edge">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Astral Email Verification</title>
        <link href="https://unpkg.com/tailwindcss@^2/dist/tailwind.min.css" rel="stylesheet">
    </head>

    <body>
        <div class="flex items-center justify-center min-h-screen p-5 bg-gray-900 min-w-screen">
            <div class="max-w-xl p-8 text-center text-gray-800 bg-white shadow-xl lg:max-w-3xl rounded-3xl lg:p-12">
                <h3 class="text-2xl font-bold">Astral Email Verification</h3>
                <div class="mt-4">
                    <a class="px-2 py-2 text-white bg-blue-600 rounded-md font-semibold" href="{{ .AuthUrl }}/integrations/email?code={{ .Code }}">Click to Verify Email</a>
                    <p class="mt-4 text-sm">If you're having trouble clicking the "Verify Email Address" button, copy
                        and
                        paste
                        the URL below
                        into your web browser:
                        <a href="{{ .AuthUrl }}/integrations/email?code={{ .Code }}" class="text-blue-600">{{ .AuthUrl }}/integrations/email?code={{ .Code }}</a>
                    </p>
                </div>
                <p class="text-sm mt-4 text-gray-600">If you didn't request this, ignore it!</p>
            </div>
        </div>

    </body>
</html>
`
