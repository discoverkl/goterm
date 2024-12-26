package term

// We should provide seperate styles for each different type of html elements.
// Except for the block style, which defines a row and a box for the content.

const BodyStyle = `
html, body {
	/* enable top level elements in body to take up the full height */
	height: 100%;
}
body {
	/* remove default margin */
	margin: 0;

	/* slightly off-white background color */
	background-color: #f8f8f8;
}
`

const IframeStyle = `
iframe {
    /* transparent background will make a iframe more like a div */
	background-color: transparent;

	/* remove border */
	border: none;
}
iframe.echart {
    /* set a minimum width for echart */
    min-width: 916px;
}
`

// Block div for html content such as charts and plots.
// For x-axis, it will take up the full width of the parent element, center it's content
// if it's smaller than the parent element, and overflow on the x-axis if it's larger.
// For y-axis, it will either use a fixed height or dynamically adjust to the content.
// Input: (x, y)
// Block: (width: 100%, height: auto|y)
// Child: (width: 100%|x, height: 100%)
const BlockStyle = `
div.goterm-row {
    /* Center the content */
    display: flex;
    justify-content: space-around;
    align-items: center;

    /* Scroll on x-axis */
    overflow-x: auto;
    overflow-y: hidden;

    /* Default background color */
    background-color: white;
}

div.goterm-box {
    /* Respect the width of the child element */
    flex: 0 0 auto;

    /* User can override the width and height */
    width: auto;
    height: auto;

    /* Center the content when it's smaller */
    display: flex;
    justify-content: left;
    align-items: center;

    /* No scrollbars */
    overflow:hidden;
}
  
div.goterm-box > :first-child {
    /* Override the width and height */
    width: 100%;
    height: 100%;
}
`

const TextStyle = `
pre.goterm {
    /* Background color similar to modern terminals */
    background-color: #1e1e1e;
    
    /* Text color with 95% brightness using HSL for modern browsers */
    color: hsl(0deg 0% 95%);
    
    /* Modern font settings */
    font-family: monaco, monospace, 'Consolas', 'Courier New';
    font-size: 1rem;
    line-height: 1.5;

	/* Remove default margin */
	margin: 0;
    
    /* Padding for better spacing */
    padding: 0.5rem;
    
    /* Border to simulate a modern terminal window */
    border: 1px solid #333;

    /* Modern text handling */
    white-space: pre-wrap;
    word-break: break-all;
    
    /* Cursor style for interactivity feel */
    cursor: text;

    /* Modern shadow for depth */
    box-shadow: 0 0 10px rgba(0, 0, 0, 0.5);
    
    /* Modern border-radius for a softer look */
    border-radius: 0.25rem;

    /* Optional: Custom scrollbar for a more native look */
    overflow-y: auto;
    scrollbar-width: thin;
    scrollbar-color: #888 #1e1e1e;
}
`

const ScrollScript = `
<script>
    let autoScroll = true;
    let lastScrollTop = 0;

    // Function to scroll to the bottom of the page
    function scrollToBottom() {
        if (autoScroll) {
			window.scrollTo({
                top: document.body.scrollHeight,
                behavior: 'smooth'
            });
        }
    }

    // Listen for scroll events to if we should stop auto-scrolling.
	// This will set autoScroll to false if the user scrolls up.
    window.addEventListener('scroll', function() {
        let st = window.pageYOffset || document.documentElement.scrollTop;
        // Check if the user is scrolling up
        if (lastScrollTop - st >= 2) {
            autoScroll = false;
        }
        lastScrollTop = st <= 0 ? 0 : st; // For Mobile or negative scrolling
    });

    // Periodically check if we should restart auto-scrolling.
	// This will set autoScroll to true if the user is at the bottom of the page.
    function checkScrollPosition() {
        if (!autoScroll && (document.body.scrollHeight - window.innerHeight - window.scrollY <= 100)) {
            autoScroll = true;
        }
    }

    // Start auto-scrolling
    setInterval(scrollToBottom, 200);

    // Start checking scroll position every second
    setInterval(checkScrollPosition, 1000);
</script>
`
