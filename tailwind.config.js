/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./templates/**/*.{html,js}",
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

