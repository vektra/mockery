Getting Started
================

Installation
-------------

### GitHub Release <small>recommended</small>

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
const start = performance.now();

function insert_installation_command(element_to_override,version){
    element_to_override.innerHTML=`
``` title=""
go install github.com/vektra/mockery/v3@${version}
```
`;
}

function compareSemver(v1, v2) {
    const parseVersion = (version) => {
        const [main, preRelease] = version.replace(/^v/, "").split("-");
        const mainParts = main.split('.').map(Number);
        const preParts = preRelease ? preRelease.split('.').map((part) => isNaN(part) ? part : Number(part)) : [];
        return { mainParts, preParts };
    };

    const compareParts = (a, b) => {
        for (let i = 0; i < Math.max(a.length, b.length); i++) {
            const partA = a[i] || 0;
            const partB = b[i] || 0;
            if (partA > partB) return 1;
            if (partA < partB) return -1;
        }
        return 0;
    };

    const { mainParts: main1, preParts: pre1 } = parseVersion(v1);
    const { mainParts: main2, preParts: pre2 } = parseVersion(v2);

    const mainComparison = compareParts(main1, main2);
    if (mainComparison !== 0) return mainComparison;

    // Compare pre-release parts
    if (pre1.length === 0 && pre2.length > 0) return 1; // No pre-release > pre-release
    if (pre1.length > 0 && pre2.length === 0) return -1; // Pre-release < no pre-release
    return compareParts(pre1, pre2);
}



const version_key="/mockery/version";
const element = document.getElementById('mockery-install-go-command');
const url = `https://api.github.com/repos/vektra/mockery/releases`;

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
      let latest_version = "";
      data.forEach((release) => {
        let release_tag=release.tag_name;
        if (!release_tag.startsWith("v3")){
          return
        };
        if (latest_version === "" || compareSemver(release_tag, latest_version) === 1) {
          latest_version=release_tag;
        };
      });
      sessionStorage.setItem(version_key, latest_version);
      insert_installation_command(element,latest_version);
    })
    .catch((error) =>{
          console.error(error);
          element.innerHTML=`failed to fetch latest release info from: https://api.github.com/repos/vektra/mockery/releases/tags/v3`;
    }
  );
}

const end = performance.now();
console.log(`Execution time for finding latest mockery tag: ${end - start} milliseconds`);
</script>
