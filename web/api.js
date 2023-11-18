let fetchedData;


function createTempDir() {
    const tempDir = fs.mkdtempSync(path.join(__dirname, 'temp-'));
    return tempDir;
}

function cloneRepository(gitUrl, targetDirectory) {
    execSync(`git clone ${gitUrl} ${targetDirectory}`);
}
async function getProjectInfo() {
    try {
        const response = await fetch('http://localhost:1323/api/github/repos');
        const data = await response.json();


        const projectSelect = document.getElementById("projectSelect");

        projectSelect.innerHTML = "";

        data.forEach(project => {
            createProjectOption(project, projectSelect);
        });

        fetchedData = data;
    } catch (error) {
        console.error('Error:', error);
    }
}

function createProjectOption(project, projectSelect) {
    const option = document.createElement("option");
    option.value = project.name;
    option.text = project.name;
    projectSelect.add(option);
}

async function getProjectData() {
    const selectedProject = document.getElementById("projectSelect").value;

    const projectInfoDiv = document.getElementById("projectInfo");
    projectInfoDiv.innerHTML = "";

    for (const project of fetchedData) {
        if (project.name === selectedProject) {
            const innerDiv = document.createElement("div");

            const responseCommits = await fetch(`http://localhost:1323/api/github/commits?projectOwner=${project.owner.login}&repoName=${project.name}`);
            const commitsData = await responseCommits.json();
            const responseLOC = await fetch('http://localhost:1323/api/github/loc?repoUrl='+project.clone_url.toString())
            const locJson = await responseLOC.json();
            let commitsMap = new Map();

            for (const commit of commitsData) {
                const newData = await fetch(commit.url, {
                    headers: {
                        'Authorization': 'Bearer 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'
                    }
                });
                const json = await newData.json();

                let userMap = commitsMap.get(json.author.login) || new Map();

                userMap.set('additions', (userMap.get('additions') || 0) + json.stats.additions);
                userMap.set('deletions', (userMap.get('deletions') || 0) + json.stats.deletions);
                userMap.set('total', (userMap.get('total') || 0) + json.stats.total);

                commitsMap.set(json.author.login, userMap);
            }

            const projectNamePara = document.createElement("p");
            projectNamePara.innerHTML = `Proje Adı: ${project.name}<br>`;
            projectNamePara.innerHTML += `Toplam Satır Sayısı: ${locJson.totalLines}<br>`;
            projectNamePara.innerHTML += `Katkı Sağlayanlar : <br>`;



            const response = await fetch(project.contributors_url.toString(), {
                headers: {
                    'Authorization': 'Bearer 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'
                }
            });
            const data = await response.json();
            data.forEach(user => {
                if (commitsMap.has(user.login)) {
                    const userMap = commitsMap.get(user.login);

                    projectNamePara.innerHTML += `<br>--- ${user.login} ---<br>`;
                    projectNamePara.innerHTML += `Additions: ${userMap.get('additions') || 0}<br>`;
                    projectNamePara.innerHTML += `Deletions: ${userMap.get('deletions') || 0}<br>`;
                    projectNamePara.innerHTML += `Total: ${userMap.get('total') || 0}<br>`;
                }
            });

            innerDiv.appendChild(projectNamePara);
            projectInfoDiv.appendChild(innerDiv);
        }
    }
}



document.addEventListener("DOMContentLoaded", function () {
    getProjectInfo();
});
