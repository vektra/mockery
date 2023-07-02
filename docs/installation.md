Getting Started
================

Installation
-------------

### GitHub Release

<small>recommended</small>

Visit the [releases page](https://github.com/vektra/mockery/releases) to download one of the pre-built binaries for your platform.

### go install

Supported, but not recommended: [see wiki page](https://github.com/vektra/mockery/wiki/Installation-Methods#go-install) and [related discussions](https://github.com/vektra/mockery/pull/456).

<div id="mockery-install-go-command"></div>

!!! warning

    Do _not_ use `@latest` as this will pull from the latest, potentially untagged, commit on master.

### Docker

Use the [Docker image](https://hub.docker.com/r/vektra/mockery)

    docker pull vektra/mockery

Generate all the mocks for your project:

	docker run -v "$PWD":/src -w /src vektra/mockery --all

### Homebrew

Install through [brew](https://brew.sh/)

    brew install mockery
    brew upgrade mockery


<script type="text/javascript">

function insert_installation_command(element_to_override,version){
    element_to_override.innerHTML=`go install github.com/vektra/mockery/v2@${version}`;
}

const version_key="/mockery/version";
const element = document.getElementById('mockery-install-go-command');
const url = `https://api.github.com/repos/vektra/mockery/releases/latest`;

let version = sessionStorage.getItem(version_key);
if (version !== null) {
    insert_installation_command(element,version);
} else {
  const request = new Request(url, {
    method: "GET",
  });

  fetch(request)
    .then((response) => response.json())
    .then((data) => {
      sessionStorage.setItem(version_key, data.name);
      insert_installation_command(element,data.name);
    })
    .catch((error) =>{
          console.error(error);
          element.innerHTML=`failed to fetch latest release info from: https://api.github.com/repos/vektra/mockery/releases/latest`;
    }
  );
}
</script>