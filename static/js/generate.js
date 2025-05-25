document.addEventListener('DOMContentLoaded', function() {
            const generateBtn = document.getElementById('generateBtn');
            const loader = document.getElementById('loader');
            const downloadMessage = document.getElementById('downloadMessage');
            const manualDownloadLink = document.getElementById('manualDownload');

            let downloadUrl = '';

            generateBtn.addEventListener('click', async function() {
                generateBtn.disabled = true;
                loader.style.display = 'block';
                downloadMessage.style.display = 'none';

                try {
                    const iframe = document.createElement('iframe');
                    iframe.style.display = 'none';
                    iframe.src = '/api/generate';
                    document.body.appendChild(iframe);

                    await new Promise(resolve => setTimeout(resolve, 3000));

                    downloadMessage.style.display = 'block';

                } catch (err) {
                    alert('Ошибка: ' + err.message);
                } finally {
                    generateBtn.disabled = false;
                    loader.style.display = 'none';
                }
            });

            manualDownloadLink.addEventListener('click', function(e) {
                e.preventDefault();
                window.location.href = '/api/generate';
            });
        });