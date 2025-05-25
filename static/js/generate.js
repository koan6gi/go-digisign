document.addEventListener('DOMContentLoaded', function() {
            const generateBtn = document.getElementById('generateBtn');
            const loader = document.getElementById('loader');
            const downloadMessage = document.getElementById('downloadMessage');
            const manualDownloadLink = document.getElementById('manualDownload');

            let downloadUrl = '';

            generateBtn.addEventListener('click', async function() {
                // Показываем индикатор загрузки
                generateBtn.disabled = true;
                loader.style.display = 'block';
                downloadMessage.style.display = 'none';

                try {
                    // Создаем скрытый iframe для скачивания
                    const iframe = document.createElement('iframe');
                    iframe.style.display = 'none';
                    iframe.src = '/api/generate';
                    document.body.appendChild(iframe);

                    // Ждем 3 секунды перед проверкой
                    await new Promise(resolve => setTimeout(resolve, 3000));

                    // Показываем сообщение об успехе
                    downloadMessage.style.display = 'block';

                } catch (err) {
                    alert('Ошибка: ' + err.message);
                } finally {
                    generateBtn.disabled = false;
                    loader.style.display = 'none';
                }
            });

            // Альтернативное скачивание
            manualDownloadLink.addEventListener('click', function(e) {
                e.preventDefault();
                window.location.href = '/api/generate';
            });
        });