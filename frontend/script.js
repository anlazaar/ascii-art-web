const titles = ["ASCII Art Generator", "Create Cool Text Art!"];
let currentTitleIndex = 0;
let currentCharIndex = 0;
let isDeleting = false;
let titleElement = document.getElementById("title");

function typeTitle() {
    const currentTitle = titles[currentTitleIndex];
    
    if (isDeleting) {
        titleElement.textContent = currentTitle.substring(0, currentCharIndex - 1);
        currentCharIndex--;
    } else {
        titleElement.textContent = currentTitle.substring(0, currentCharIndex + 1);
        currentCharIndex++;
    }

    if (!isDeleting && currentCharIndex === currentTitle.length) {
        setTimeout(() => isDeleting = true, 1000);
    } else if (isDeleting && currentCharIndex === 0) {
        isDeleting = false;
        currentTitleIndex = (currentTitleIndex + 1) % titles.length;
    }

    const typingSpeed = isDeleting ? 100 : 200;
    setTimeout(typeTitle, typingSpeed);
}

// Start the typing animation
typeTitle();

async function generateAscii() {
    const text = document.getElementById("textInput").value;
    const style = document.getElementById("styleSelect").value;
    
    try {
        const response = await fetch('http://localhost:8080/generate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                text: text,
                style: style
            })
        });

        const data = await response.json();
        
        if (data.error) {
            document.getElementById("asciiResult").textContent = `Error: ${data.error}`;
            document.getElementById("downloadBtn").disabled = true;
        } else {
            document.getElementById("asciiResult").textContent = data.art;
            document.getElementById("downloadBtn").disabled = false;
        }
    } catch (error) {
        document.getElementById("asciiResult").textContent = `Error connecting to server: ${error.message}`;
        document.getElementById("downloadBtn").disabled = true;
    }
}

function downloadAsciiArt() {
    const asciiArt = document.getElementById("asciiResult").textContent;
    const blob = new Blob([asciiArt], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    
    const a = document.createElement('a');
    a.href = url;
    a.download = 'ascii_art.txt';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
}