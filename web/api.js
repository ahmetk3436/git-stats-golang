// Global array to store fetched project data (list of repositories).
// This data is populated by getProjectInfo and used by getProjectData.
let fetchedData = [];

// API_BASE_URL is the base URL for all backend API calls.
// Assumes the backend is running on localhost:1323.
// Adjust this if your backend is hosted elsewhere.
const API_BASE_URL = 'http://localhost:1323/api';

/**
 * Fetches the list of available GitHub repositories from the backend.
 * Populates the project selection dropdown (<select id="projectSelect">) with these repositories.
 * Stores the fetched repository list in the global `fetchedData` array.
 */
async function getProjectInfo() {
    console.log("Fetching project information...");
    try {
        // Fetch repositories from the backend's GitHub proxy endpoint.
        // The backend handles authentication with GitHub.
        const response = await fetch(`${API_BASE_URL}/github/repos`);
        if (!response.ok) {
            // If the response is not OK (e.g., 4xx, 5xx errors), throw an error to be caught by the catch block.
            throw new Error(`Failed to fetch repositories: ${response.status} ${response.statusText}`);
        }
        const data = await response.json(); // Expects an array of common_types.Repository.

        const projectSelect = document.getElementById("projectSelect");
        projectSelect.innerHTML = ""; // Clear any existing options from the dropdown.

        // Create and append a new <option> for each fetched project.
        data.forEach(project => {
            createProjectOption(project, projectSelect);
        });

        fetchedData = data; // Cache the fetched repository list for later use by getProjectData.
        console.log("Project information fetched and dropdown populated.");
    } catch (error) {
        console.error('Error in getProjectInfo:', error);
        // Display a user-friendly error message in the UI.
        const projectInfoDiv = document.getElementById("projectInfo");
        if (projectInfoDiv) {
            projectInfoDiv.innerHTML = `<p>Error loading project list: ${error.message}. Please ensure the backend is running and accessible at ${API_BASE_URL}.</p>`;
        }
    }
}

/**
 * Creates an <option> element for a project and adds it to the given <select> element.
 * @param {object} project - The project object, expected to conform to common_types.Repository.
 * @param {HTMLSelectElement} projectSelect - The <select> element to which the option will be added.
 */
function createProjectOption(project, projectSelect) {
    const option = document.createElement("option");
    option.value = project.name; // Use project name as the value for selection.
    option.text = project.name;  // Display project name in the dropdown.
    projectSelect.add(option);
}

/**
 * Fetches and displays detailed data for the currently selected project from the dropdown.
 * This includes commits, lines of code (LOC), and contributor statistics.
 * All data is fetched through backend proxy endpoints, ensuring no direct calls to Git providers from the frontend.
 */
async function getProjectData() {
    const selectedProjectName = document.getElementById("projectSelect").value;
    const projectInfoDiv = document.getElementById("projectInfo");
    
    // Clear previous project data and show a loading message.
    projectInfoDiv.innerHTML = "<p>Proje verileri yükleniyor...</p>";

    // Find the selected project's details from the cached `fetchedData`.
    const selectedProject = fetchedData.find(p => p.name === selectedProjectName);
    if (!selectedProject) {
        console.error('Selected project not found in fetched data cache.');
        projectInfoDiv.innerHTML = "<p>Hata: Seçilen proje verisi bulunamadı. Lütfen listeyi yenileyin.</p>";
        return;
    }

    console.log(`Fetching data for project: ${selectedProjectName}`);

    try {
        // Extract necessary identifiers from the selected project object.
        // These fields are based on the common_types.Repository structure.
        const projectOwner = selectedProject.owner;
        const projectName = selectedProject.name;
        const projectCloneURL = selectedProject.cloneURL;

        // Fetch commits for the selected project via the backend.
        // The backend proxies this request to the appropriate Git provider.
        const responseCommits = await fetch(`${API_BASE_URL}/github/commits?projectOwner=${projectOwner}&repoName=${projectName}`);
        if (!responseCommits.ok) {
            throw new Error(`Commit verileri alınamadı: ${responseCommits.status} ${responseCommits.statusText}`);
        }
        const commitsData = await responseCommits.json(); // Expects []*common_types.Commit

        // Fetch Lines of Code (LOC) for the selected project via the backend.
        const responseLOC = await fetch(`${API_BASE_URL}/github/loc?repoUrl=${projectCloneURL}`);
        if (!responseLOC.ok) {
            throw new Error(`LOC verileri alınamadı: ${responseLOC.status} ${responseLOC.statusText}`);
        }
        const locJson = await responseLOC.json(); // Expects { totalLines: number }

        // Aggregate commit statistics (additions, deletions, total changes, commit count) by author.
        let commitsMap = new Map(); // Using a Map to store aggregated stats per author.
        for (const commit of commitsData) { // `commit` is of type common_types.Commit
            // Use commit.author.name as the primary key for aggregation.
            // Fallback to email or "Unknown Author" if name is not available.
            // Consistency in how authors are identified across commit data and contributor data is important.
            let authorKey = commit.author.name || commit.author.email || "Unknown Author";

            let userStats = commitsMap.get(authorKey) || { additions: 0, deletions: 0, total: 0, commitCount: 0 };
            userStats.additions += commit.stats.additions;
            userStats.deletions += commit.stats.deletions;
            userStats.total += commit.stats.total;
            userStats.commitCount += 1; // Increment commit count for this author.
            commitsMap.set(authorKey, userStats);
        }

        // Fetch contributors for the selected project via the backend.
        const contributorsResponse = await fetch(`${API_BASE_URL}/github/contributors?owner=${projectOwner}&repoName=${projectName}`);
        if (!contributorsResponse.ok) {
            throw new Error(`Katkıda bulunanlar alınamadı: ${contributorsResponse.status} ${contributorsResponse.statusText}`);
        }
        const contributorsData = await contributorsResponse.json(); // Expects []*common_types.User

        // --- Display project information and contributor stats in the UI ---
        const projectDetailsHtml = `
            <p><strong>Proje Adı:</strong> ${projectName}</p>
            <p><strong>Toplam Satır Sayısı:</strong> ${locJson.totalLines !== undefined ? locJson.totalLines : 'Hesaplanamadı'}</p>
            <p><strong>Katkı Sağlayanlar:</strong></p>
        `;
        projectInfoDiv.innerHTML = projectDetailsHtml; // Set initial project details.

        const contributorsListElement = document.createElement("ul"); // Create a list for contributors.
        contributorsData.forEach(contributor => { // `contributor` is of type common_types.User
            // Display key for the contributor, preferring login, then name.
            let contributorDisplayKey = contributor.login || contributor.name || "Unknown Contributor";
            
            // Attempt to find matching commit stats for this contributor from the aggregated commitsMap.
            // This matching relies on the identifier used for `authorKey` above (name/email from commit)
            // aligning with `contributor.login` or `contributor.name`.
            let userCommitStats = commitsMap.get(contributor.login) || commitsMap.get(contributor.name);
            // If contributor.email is available and was used as authorKey, that could be another fallback.

            const listItem = document.createElement("li");
            listItem.innerHTML = `--- <strong>${contributorDisplayKey}</strong> ---<br>`;
            if (userCommitStats) {
                listItem.innerHTML += `Commits: ${userCommitStats.commitCount}<br>`;
                listItem.innerHTML += `Additions (Satır): ${userCommitStats.additions}<br>`;
                listItem.innerHTML += `Deletions (Satır): ${userCommitStats.deletions}<br>`;
                listItem.innerHTML += `Total Changes (Satır): ${userCommitStats.total}<br>`;
            } else {
                // If no commit stats are found for this contributor (e.g., they made no commits or matching failed).
                listItem.innerHTML += `<em>(Bu kullanıcı için commit istatistikleri bulunamadı veya commit yapmamış.)</em><br>`;
            }
            listItem.innerHTML += `<br>`; // Add some spacing between contributors.
            contributorsListElement.appendChild(listItem);
        });
        projectInfoDiv.appendChild(contributorsListElement); // Add the list of contributors to the display area.
        console.log("Project data displayed successfully.");

    } catch (error) {
        console.error('Error in getProjectData:', error);
        projectInfoDiv.innerHTML = `<p>Proje detayları yüklenirken hata oluştu: ${error.message}. Lütfen backend'in çalıştığından ve ilgili API endpointlerinin erişilebilir olduğundan emin olun.</p>`;
    }
}

/**
 * Event listener for when the DOM content is fully loaded.
 * Initializes the application by fetching the project list to populate the dropdown.
 */
document.addEventListener("DOMContentLoaded", function () {
    console.log("DOM fully loaded and parsed. Initializing application.");
    // The application no longer needs to fetch a configuration containing an API token.
    // Directly call getProjectInfo to populate the project list.
    getProjectInfo();
});
