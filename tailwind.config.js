/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./templates/**/*.{html,js}",
    ],
    theme: {
        extend: {
            fontFamily: {
                sans: ['"VT323"', 'monospace']
            },
        },
    },
    plugins: [
        function ({ addVariant }) {
            addVariant('child', '& > *');
        }
    ],
}

