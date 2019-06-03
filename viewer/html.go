package main

var htmlPage []byte = []byte(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width,initial-scale=1" />
	<title>Data Viewer</title>
	<style>
		body {
			font-family: monospace;
		}

		header {
			display: flex;
			font-size: 0.6rem;
			border-bottom: #444 1px solid;
		}
	</style>
	<script>
	(function() {
		function renderItem(item) {
			const el = document.createElement('pre');
			el.innerText = JSON.stringify(item, null, 2);

			return el;
		}

		async function update() {
			const main = document.querySelector('main');
			const status = document.querySelector('#status');
			
			main.innerHTML = '';

			try {
				status.innerText = 'Loading...';

				const data = await fetch("/collected");
				if(!data.ok) throw data;

				const json = await data.json();

				if(json && json.forEach) {
					json.forEach(item => {
						main.appendChild( renderItem( item ));
					});
				}

				status.innerText = 'Loaded';
			} catch(err) {
				status.innerText = 'Error -- check your console.';
				console.error('Failed to load the data list');
				console.error(err);
			}
		}

		function main() {
			// Set interval
			update();
			setInterval(update, 3000);
		}

		document.addEventListener("DOMContentLoaded", main);
	})();
	</script>
</head>
<body>
	<h1>Data Viewer</h1>
	<header>
		<div id="status"></div>
	</header>
	<main></main>
</body>
</html>
`)
