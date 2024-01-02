/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./templates/**/*.{html,js,templ}",
    ],
    theme: {
        extend: {
            fontFamily: {
                sans: ['"JetBrains Mono"', 'monospace'],

            },
        },
    },
    plugins: [
        function ({ addVariant }) {
            addVariant('child', '& > *');
        }
    ],
}

