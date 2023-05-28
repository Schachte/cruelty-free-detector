const searchBox = document.getElementById('search-box');
const evaluateButton = document.getElementById('evaluate-button');
const responseContainer = document.getElementById('response-container');

evaluateButton.addEventListener('click', () => {
    const query = searchBox.value;

    evaluateButton.setAttribute('disabled', 'disabled');
    evaluateButton.innerHTML = '<div class="flex w-full justify-center items-center space-x-3"><img src="assets/spinner3.gif"/><span>Evaluating cruelty, please wait</span></div>';

    fetch('http://tuffpuff.com:8080', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ query }),
    })
    .then(response => response.json())
    .then(data => {
        if (!data.cruelty_free) {
            let responseText = `The company <u>${data.company_name}</u> is not cruelty free. <br/> `;
            if (data.parent_company) {
                responseText += `${data.company_name} is owned by <b>${data.parent_company}</b>. `;
            }

            if (data.offenses && data.offenses.length > 0) {
                if (data.offenses.length > 0) {
                    responseText += "<br/><br>Known offenses include:</b>";
                    responseText += "<ul>";
                    data.offenses.forEach(offense => {
                        responseText += `<li> -${offense}</li>`;
                    });
                    responseText += "</ul>";
                }
            }

            if (data.alternatives && data.alternatives.length > 0) {
                if (data.alternatives.length > 0) {
                    responseText += "<br/><b>Alternatives include:</b>";
                    responseText += "<ul>";
                    data.alternatives.forEach(alternative => {
                        responseText += `<li> -${alternative.company_name || alternative}</li>`;
                    });
                    responseText += "</ul>";
                }
            }

            responseContainer.innerHTML = responseText;
            responseContainer.classList.remove('hidden', 'bg-green-500');
            responseContainer.classList.add('bg-red-500', 'text-white');
        } else {
            responseContainer.textContent = `The company ${data.company_name} is cruelty free!`;
            responseContainer.classList.remove('hidden', 'bg-red-500');
            responseContainer.classList.add('bg-green-500', 'text-white');
        }

        evaluateButton.textContent = 'Evaluate Product or Company';
        evaluateButton.removeAttribute('disabled');
    })
    .catch((error) => {
        console.error('Error:', error);
    });
});
