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
    // console.log("Fetching project information..."); // Retained for initial load feedback, can be removed later.
    const projectSelect = document.getElementById("projectSelect");
    const projectInfoDiv = document.getElementById("projectInfo"); // For error display

    try {
        // Fetch repositories from the backend's GitHub proxy endpoint.
        const response = await fetch(`${API_BASE_URL}/github/repos`);
        if (!response.ok) {
            throw new Error(`Failed to fetch repositories: ${response.status} ${response.statusText}`);
        }
        const data = await response.json(); // Expects an array of common_types.Repository.

        projectSelect.innerHTML = ""; // Clear "Loading projects..." or previous options.

        // Add a default, non-selectable "Select a project..." option.
        const defaultOption = document.createElement("option");
        defaultOption.value = "";
        defaultOption.textContent = "Select a project...";
        projectSelect.appendChild(defaultOption);

        // Create and append a new <option> for each fetched project.
        data.forEach(project => {
            createProjectOption(project, projectSelect);
        });

        fetchedData = data; // Cache the fetched repository list.
        // console.log("Project information fetched and dropdown populated."); // Debug log
    } catch (error) {
        console.error('Error in getProjectInfo:', error);
        if (projectInfoDiv) {
            projectInfoDiv.innerHTML = `<p class="error-message">Error loading project list: ${error.message}. Please ensure the backend is running and accessible at ${API_BASE_URL}.</p>`;
        }
        if (projectSelect) {
            projectSelect.innerHTML = '<option value="">Could not load projects</option>';
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
    option.value = project.name;
    option.text = project.name;
    projectSelect.add(option);
}

/**
 * Fetches and displays detailed data for the currently selected project.
 * Uses DOM manipulation to build the display for better structure and styling.
 * Implements loading and error feedback within the `projectInfoDiv`.
 */
async function getProjectData() {
    const selectedProjectName = document.getElementById("projectSelect").value;
    const projectInfoDiv = document.getElementById("projectInfo");

    // Clear previous project data and show a loading message.
    projectInfoDiv.innerHTML = `<p class="loading-message">Loading project data for '${selectedProjectName}'...</p>`;

    const selectedProject = fetchedData.find(p => p.name === selectedProjectName);
    if (!selectedProject) {
        console.error('Selected project not found in fetched data cache.');
        projectInfoDiv.innerHTML = "<p class='error-message'>Error: Selected project data could not be found. Please refresh.</p>";
        return;
    }

    // console.log(`Fetching data for project: ${selectedProjectName}`); // Debug log

    try {
        const projectOwner = selectedProject.owner;
        const projectName = selectedProject.name;
        const projectCloneURL = selectedProject.cloneURL;

        // Fetch commits
        const responseCommits = await fetch(`${API_BASE_URL}/github/commits?projectOwner=${projectOwner}&repoName=${projectName}`);
        if (!responseCommits.ok) throw new Error(`Commit data could not be retrieved: ${responseCommits.statusText}`);
        const commitsData = await responseCommits.json();

        // Fetch LOC
        const responseLOC = await fetch(`${API_BASE_URL}/github/loc?repoUrl=${projectCloneURL}`);
        if (!responseLOC.ok) throw new Error(`LOC data could not be retrieved: ${responseLOC.statusText}`);
        const locJson = await responseLOC.json();

        // Aggregate commit stats
        let commitsMap = new Map();
        for (const commit of commitsData) {
            let authorKey = commit.author.name || commit.author.email || "Unknown Author";
            let userStats = commitsMap.get(authorKey) || { additions: 0, deletions: 0, total: 0, commitCount: 0 };
            userStats.additions += commit.stats.additions;
            userStats.deletions += commit.stats.deletions;
            userStats.total += commit.stats.total;
            userStats.commitCount += 1;
            commitsMap.set(authorKey, userStats);
        }

        // Fetch contributors
        const contributorsResponse = await fetch(`${API_BASE_URL}/github/contributors?owner=${projectOwner}&repoName=${projectName}`);
        if (!contributorsResponse.ok) throw new Error(`Contributor data could not be retrieved: ${contributorsResponse.statusText}`);
        const contributorsData = await contributorsResponse.json();

        // --- Build UI with DOM elements ---
        projectInfoDiv.innerHTML = ''; // Clear loading message

        const nameHeader = document.createElement('h2');
        nameHeader.textContent = projectName;
        projectInfoDiv.appendChild(nameHeader);

        const locP = document.createElement('p');
        locP.innerHTML = `<strong>Total Lines of Code:</strong> ${locJson.totalLines !== undefined ? locJson.totalLines : 'N/A'}`;
        projectInfoDiv.appendChild(locP);

        const contributorsHeader = document.createElement('h3');
        contributorsHeader.textContent = "Contributors";
        projectInfoDiv.appendChild(contributorsHeader);

        if (contributorsData.length === 0) {
            const noContributorsP = document.createElement('p');
            noContributorsP.textContent = "No contributor data available for this project.";
            projectInfoDiv.appendChild(noContributorsP);
        } else {
            const contributorsUl = document.createElement('ul');
            contributorsData.forEach(contributor => {
                const li = document.createElement('li');
                // ClassName for styling individual contributor items from CSS
                li.className = 'contributor-item'; 
                const contributorKey = contributor.login || contributor.name || "Unknown Contributor";
                let statsHtml = `<strong>${contributorKey}</strong>`;
                
                const userCommitStats = commitsMap.get(contributor.login) || commitsMap.get(contributor.name);
                if (userCommitStats) {
                    statsHtml += `<p>Commits: ${userCommitStats.commitCount}</p>`;
                    statsHtml += `<p>Additions: ${userCommitStats.additions}</p>`;
                    statsHtml += `<p>Deletions: ${userCommitStats.deletions}</p>`;
                    statsHtml += `<p>Total Changes (Lines): ${userCommitStats.total}</p>`;
                } else {
                    statsHtml += `<p><em>(No specific commit stats found for this contributor by name/login match)</em></p>`;
                }
                li.innerHTML = statsHtml;
                contributorsUl.appendChild(li);
            });
            projectInfoDiv.appendChild(contributorsUl);
        }
        // console.log("Project data displayed successfully using DOM manipulation."); // Debug log

    } catch (error) {
        console.error('Error in getProjectData:', error);
        projectInfoDiv.innerHTML = `<p class="error-message">Failed to load project details: ${error.message}. Please try again or select another project.</p>`;
    }
}

/**
 * Event listener for when the DOM content is fully loaded.
 * Initializes the application by fetching the project list and setting up event listeners.
 */
document.addEventListener("DOMContentLoaded", function () {
    // console.log("DOM fully loaded and parsed. Initializing application."); // Debug log
    getProjectInfo(); // Populate the project dropdown

    const projectSelectElement = document.getElementById("projectSelect");
    if (projectSelectElement) {
        projectSelectElement.addEventListener("change", function() {
            const selectedProjectName = this.value;
            // console.log(`Project selection changed to: ${selectedProjectName}`); // Debug log
            if (selectedProjectName && selectedProjectName !== "") {
                getProjectData();
            } else {
                const projectInfoDiv = document.getElementById("projectInfo");
                if (projectInfoDiv) {
                    projectInfoDiv.innerHTML = "<p>Please select a project to see its information.</p>";
                }
            }
        });
    } else {
        console.error("Error: Project select dropdown element with ID 'projectSelect' not found.");
    }
});
