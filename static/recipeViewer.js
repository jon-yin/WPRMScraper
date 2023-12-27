const iframe = document.querySelector("#recipe-detail");

let siteData;

const recipesById = new Map();
const recipeSortOptions = {
    "name": (recipe1, recipe2) => {
        return recipe1.Name.localeCompare(recipe2.Name);
    },
    // rate highest rated to lowest
    "rating": (recipe1, recipe2) => {
        if (Math.abs(recipe1.Rating - recipe2.Rating) < 0.001) {
            // if relatively equal, rate by number rated
            return recipeSortOptions["numRated"](recipe1, recipe2);
        }
        if (recipe1.Rating < recipe2.Rating) {
            return 1;
        }
        return -1;
    },
    "numRated": (recipe1, recipe2) => {
        if (recipe1.NumRated < recipe2.NumRated) {
            return 1;
        }
        if (recipe1.NumRated === recipe2.NumRated) {
            return 0;
        }
        return -1;
    }
}

const nonAlphaNumericRegexp = /\W/g;
const spaceRegex = /\s+/g;

function indexRecipesById(recipes) {
    recipes.forEach(recipe => {
        recipesById.set(recipe.ID, recipe);
    })
}

function categorizeRecipes(recipes) {
    const cuisinesMap = {};
    const coursesMap = {};
    recipes.forEach(recipe => {
        recipe.Cuisine?.forEach(cuisine => {
            const cuisineKey = idSafe(cuisine);
            if (!cuisinesMap[cuisineKey]) {
                cuisinesMap[cuisineKey] = {
                    recipes: [recipe.ID],
                    name: cuisine
                };
            } else {
                cuisinesMap[cuisineKey].recipes.push(recipe.ID);
            }
        });
        siteData.cuisines = cuisinesMap;

        recipe.Course?.forEach(course => {
            const courseKey = idSafe(course);
            if (!coursesMap[courseKey]) {
                coursesMap[courseKey] ={ 
                    recipes: [recipe.ID],
                    name: course
                };
            } else {
                coursesMap[courseKey].recipes.push(recipe.ID);
            }
        })
        siteData.courses = coursesMap;
        siteData.allRecipes.push(recipe.ID);
    })
}

function createElement(specs) {
    const { id, attrs, classes, children, type, onClick } = specs
    const element = document.createElement(type);
    if (attrs) {
        Object.entries(attrs).forEach(([attrKey, attrVal]) => {
            element.setAttribute(attrKey, attrVal);
        });
    }
    if (classes) {
        element.classList.add(...classes)
    }
    if (id) {
        element.id = id;
    }
    if (onClick) {
        element.addEventListener("click", onClick);
    }
    const convertedChildren = children?.map(child => {
        if (typeof child === "string") {
            return child;
        }
        if (child instanceof Element) {
            return child;
        }
        return createElement(child);
    }) ?? [];
    element.append(...convertedChildren);
    return element;
}

function attachMenubarControls() {
    const expandIcon = document.querySelector("#expand");
    expandIcon.addEventListener("click", () => {
        document.querySelectorAll(".collapse").forEach(elem => {
            const collapse = bootstrap.Collapse.getOrCreateInstance(elem, { toggle: false });
            collapse.show();
        });
    });
    const collapseIcon = document.querySelector("#collapse");
    collapseIcon.addEventListener("click", () => {
        document.querySelectorAll(".collapse").forEach(elem => {
            const collapse = bootstrap.Collapse.getOrCreateInstance(elem, { toggle: false });
            collapse.hide();
        });
    })

    const searchField = document.querySelector("#search");
    searchField.value = "";
    const clearIcon = document.querySelector("#clear");
    clearIcon.addEventListener("click", () => {
        searchField.value = "";
    });
    const fireSearchQuery = searchTerm => {
        const normTerm = normalizeString(searchTerm);
        recipesById.forEach((v) => {
            const normalName = normalizeString(v.Name);
            v.isVisible = normalName.includes(normTerm) || v.Keywords.some(tag => tag.includes(normTerm)) || v.Ingredients.some(ingredient => ingredient.includes(normTerm));
        });
        buildRecipeLists();
        updateRecipeCounts();
    }
    searchField.addEventListener("input", e => {
        fireSearchQuery(e.target.value);
    })

    const lightModeIcon = document.querySelector("#light-mode");
    const darkModeIcon = document.querySelector("#dark-mode");
    lightModeIcon.addEventListener("click", () => {
        siteData.isDarkMode = false;
        document.querySelector("html").setAttribute("data-bs-theme", "light");
        darkModeIcon.classList.remove("d-none");
        lightModeIcon.classList.add("d-none");
    });
    darkModeIcon.addEventListener("click", () => {
        siteData.isDarkMode = true;
        document.querySelector("html").setAttribute("data-bs-theme", "dark");
        lightModeIcon.classList.remove("d-none");
        darkModeIcon.classList.add("d-none");
    });

    if (siteData.isDarkMode) {
        darkModeIcon.classList.add("d-none");
    } else {
        lightModeIcon.classList.add("d-none");
    }

    const dropdownButton = document.querySelector("#dropdownText");
    dropdownButton.innerText = document.querySelector(`[data-sort="${siteData.selectedSort}"]`).innerText;
    const sortMenu = document.querySelector("#sortMenu");
    sortMenu.addEventListener("click", e => {
        if (e.target.tagName !== "A") {
            return;
        }
        siteData.selectedSort = e.target.getAttribute("data-sort");
        dropdownButton.innerText = e.target.innerText;
        buildRecipeLists();
    });
}

function listGroupClickListener(e) {
    if (e.target.classList.contains("list-group-item")) {
        iframe.src = "./recipes/" + e.target.getAttribute("data-uuid") + ".html";
    }
}

function buildChildAccordion(recipeGroups) {
    const recipeTuples = Object.entries(recipeGroups).sort(([,{name: name1}], [, {name: name2}]) => {
        return name1.localeCompare(name2);
    });
    const accordionItems = recipeTuples.map(([key, recipeObj]) => {
        const { name } = recipeObj;
        const accordBodyId = `${key}-accord-body`;
        const listGroupId = `${key}-list-group`;
        const accordItem = createElement({
            type: "div",
            classes: ["accordion-item"],
            children: [
                // accordion header
                {
                    type: "h3",
                    classes: ['accordion-header'],
                    children: [
                        {
                            type: "button",
                            children: [name.charAt(0).toUpperCase() + name.slice(1), createElement({
                                type: "span",
                                classes: ["recipe-count"],
                            })],
                            attrs: {
                                "type": "button",
                                "data-bs-target": `#${accordBodyId}`,
                                "data-bs-toggle": "collapse",
                                "aria-expanded": "false",
                                "aria-controls": accordBodyId
                            },
                            classes: ["accordion-button", "collapsed"]
                        }
                    ]
                },
                // accordion body
                {
                    type: "div",
                    id: accordBodyId,
                    classes: ["accordion-collapse", "collapse"],
                    children: [
                        {
                            type: "div",
                            classes: ["accordion-body"],
                            children: [createElement({
                                type: "div",
                                classes: ["list-group"],
                                id: listGroupId,
                                onClick: listGroupClickListener
                            })]
                        }
                    ]
                }
            ],
        });
        return accordItem;
    })

    const accordion = createElement({
        type: "div",
        classes: ["accordion"],
        children: accordionItems
    });
    return accordion;
}

function buildRecipeLists() {

    // build list groups first
    const buildListGroup = (parentContainer, recipeIds) => {
        const listGroupItems = recipeIds
        .map(id => recipesById.get(id))
        .filter(recipe => recipe.isVisible)
        .sort(recipeSortOptions[siteData.selectedSort])
        .map(recipe => recipe.listGroupItem.cloneNode(true));
        parentContainer.replaceChildren(...listGroupItems);
    }
    Object.entries(siteData.cuisines).forEach(([cuisine, {recipes: recipeIds}]) => {
        
        buildListGroup(document.querySelector(`#${cuisine}-list-group`), recipeIds) });

    Object.entries(siteData.courses).forEach(([course, {recipes: recipeIds}]) => 
        buildListGroup(document.querySelector(`#${course}-list-group`), recipeIds));

    buildListGroup(document.querySelector('#all-list-group'), siteData.allRecipes);
}

function updateRecipeCounts() {
    document.querySelectorAll(".recipe-count").forEach(e => {
        const accordionBody = document.querySelector(e.getAttribute("data-recipe-group"));
        e.innerText = `(${accordionBody.querySelectorAll(".list-group-item").length})`;
    })
}

function writeRecipeCounts() {
    const listGroupCount = 2;
    document.querySelectorAll(".recipe-count").forEach((e, i) => {
        if (e.hasAttribute("data-recipe-group")) {
            return;
        }
        e.setAttribute("data-recipe-group", `#recipe-group-${listGroupCount + i}`);
        const accordBody = e.parentNode.parentNode.parentNode.querySelector(".accordion-body");
        accordBody.id = `recipe-group-${listGroupCount + i}`;
    });
    updateRecipeCounts();
}

function attachToAccordion(parentItem, childAccordion) {
    const bodyItem = parentItem.querySelector(".accordion-body");
    bodyItem.append(childAccordion);
}

function indexRecipes(recipes) {
    indexRecipesById(recipes);
    categorizeRecipes(recipes);
}

function buildSite() {
    const courseParentAccordion = document.querySelector("#byCourseAccordion")
    const courseChildAccordion = buildChildAccordion(siteData.courses);

    const cuisinesParentAccordion = document.querySelector("#byCuisineAccordion");
    const cuisinesChildAccordion = buildChildAccordion(siteData.cuisines);

    const allRecipesAccordion = document.querySelector("#allRecipesAccordion");
    const allRecipesList = createElement({
        type: "div",
        classes: ["list-group"],
        id: "all-list-group",
        onClick: listGroupClickListener
    });

    // accordion stuff
    attachToAccordion(courseParentAccordion, courseChildAccordion);
    attachToAccordion(cuisinesParentAccordion, cuisinesChildAccordion);
    attachToAccordion(allRecipesAccordion, allRecipesList);
    buildRecipeLists();
    writeRecipeCounts();

    // header controls
    const dimen = "48"; // (double the default icon size)
    feather.replace({ height: dimen, width: dimen});
    attachMenubarControls();
}

function idSafe(str) {
    const normalized = normalizeString(str);
    if (/^[^a-zA-Z]/.test(normalized)) {
        return "z" + normalized;
    }
    return normalized;
}

function normalizeString(str, preserveSpaces = false) {
    const latinized = str.normalize("NFD").replace(/[\u0300-\u036f]/g, "");
    let resString = latinized;
    if (preserveSpaces) {
        resString = resString.replaceAll(spaceRegex, "_");
    }
    resString = resString.replaceAll(nonAlphaNumericRegexp, "");
    return resString.toLowerCase();
}

function augmentRecipeData(recipeData) {
    return recipeData.map(recipe => {
        const clonedRecipe = {
            ...recipe
        };
        clonedRecipe.cleanName = normalizeString(recipe.Name);
        clonedRecipe.Keywords = recipe.Keywords?.map(normalizeString) ?? [];
        clonedRecipe.Ingredients = recipe.Keywords?.map(normalizeString) ?? [];
        clonedRecipe.isVisible = true;
        clonedRecipe.listGroupItem = createElement({
            type: "a",
            children: [`${clonedRecipe.Name} - ${clonedRecipe.Rating}★`],
            attrs: {
                "data-toggle": "list",
                "href": "#",
                "data-uuid": clonedRecipe.ID,
            },
            classes: ["list-group-item", "list-group-item-action", "list-group-item-light"]
        });
        return clonedRecipe;
    });
}

function initialize() {
    annotatedRecipeData = augmentRecipeData(recipeData);
    iframe.src = "";
    siteData = {
        courses: {
        },
        cuisines: {
        },
        allRecipes: [],
        isDarkMode: false,
        selectedSort: "name",
        pendingSearch: null,
        currentSearch: ""
    }
    indexRecipes(annotatedRecipeData);
    buildSite();
}

initialize();