const iframe = document.querySelector("#recipe-detail");

let siteData;
let selectedSort = "name";

const recipesById = new Map();
const recipeSortOptions = {
    "name": (id1, id2) => {
        const recipe1 = recipesById.get(id1);
        const recipe2 = recipesById.get(id2);
        return recipe1.Name.localeCompare(recipe2.Name);
    },
    "rating": (id1, id2) => {
        const recipe1 = recipesById.get(id1);
        const recipe2 = recipesById.get(id2);
        if (Math.abs(recipe1.Rating - recipe2.Rating) < 0.001) {
            return 0;
        }
        if (recipe1.Rating < recipe2.Rating) {
            return -1;
        }
        return 1;
    },
    "numRated": (id1, id2) => {
        const recipe1 = recipesById.get(id1);
        const recipe2 = recipesById.get(id2);
        if (recipe1.NumRated < recipe2.NumRated) {
            return -1;
        }
        if (recipe1.NumRated === recipe2.NumRated) {
            return 0;
        }
        return 1;
    }
}

const nonAlphaNumericRegexp = /[^a-zA-Z0-9]/g;

function applySort() {
    const curSort = recipeSortOptions[selectedSort];
    const sortListGroup = listGroup => {
        const childrenArr = Array.from(listGroup.children);
        childrenArr.sort((elem1, elem2) => {
            return curSort(elem1.getAttribute("data-uuid"), elem2.getAttribute("data-uuid"));
        });
        listGroup.replaceChildren(...childrenArr);
    }

    siteData.courses.listGroups.forEach(sortListGroup);
    siteData.cuisines.listGroups.forEach(sortListGroup);
    sortListGroup(siteData.allRecipes.listGroup);
}

function makeIdSafe(str) {
    //trim non-alphanumeric characters
    return str.replaceAll(/[^a-zA-Z0-9]/g, "");
}

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
            if (!cuisinesMap[cuisine]) {
                cuisinesMap[cuisine] = [recipe.ID];
            } else {
                cuisinesMap[cuisine].push(recipe.ID);
            }
        });
        // a tuple [cuisines, [recipes]]
        siteData.cuisines.recipes = Object.keys(cuisinesMap).sort().map(k => [k, cuisinesMap[k]]);

        recipe.Course?.forEach(course => {
            if (!coursesMap[course]) {
                coursesMap[course] = [recipe.ID];
            } else {
                coursesMap[course].push(recipe.ID);
            }
        })

        siteData.courses.recipes = Object.keys(coursesMap).sort().map(k => [k, coursesMap[k]]);
        siteData.allRecipes.recipes.push(recipe);
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

function attachIconControls() {
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
        console.log(searchTerm);
    }
    let timeoutId;
    searchField.addEventListener("input", e => {
        clearTimeout(timeoutId);
        timeoutId = setTimeout(fireSearchQuery ,1000, e.target.value);
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
    
}

function buildRecipeList(recipeIds) {
    return createElement({
        type: "div",
        classes: ["list-group"],
        children: recipeIds.map(recipeId => {
            const recipe = recipesById.get(recipeId);
            return {
            type: "a",
            children: [recipe.Name],
            attrs: {
                "data-toggle": "list",
                "href": "#",
                "data-uuid": recipeId
            },
            classes: ["list-group-item", "list-group-item-action", "list-group-item-light"],
            onClick: (e) => {
                // do not scroll screen after clicking on a item
                e.preventDefault();
                // set iframe to display new image
                const id = e.target.getAttribute("data-uuid");
                iframe.src = "./recipes/" + id + ".html";
            }
    }})});
}

function buildChildAccordion(courseTuple) {
    const accordionItems = courseTuple.map(([key, recipes]) => {
        const id = makeIdSafe(key);
        const accordBodyId = `${id}-body`;
        const accordItem = createElement({
            type: "div",
            classes: ["accordion-item"],
            id,
            children: [
                // accordion header
                {
                    type: "h3",
                    classes: ['accordion-header'],
                    children: [
                        {
                            type: "button",
                            children: [key.charAt(0).toUpperCase() + key.slice(1), createElement({
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
                            // list recipes in each sub category
                            children: [buildRecipeList(recipes)]
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
    const courseChildAccordion = buildChildAccordion(siteData.courses.recipes);
    siteData.courses.listGroups = courseChildAccordion.querySelectorAll(".list-group");

    const cuisinesParentAccordion = document.querySelector("#byCuisineAccordion");
    const cuisinesChildAccordion = buildChildAccordion(siteData.cuisines.recipes);
    siteData.cuisines.listGroups = cuisinesChildAccordion.querySelectorAll(".list-group");

    const allRecipesAccordion = document.querySelector("#allRecipesAccordion");
    const allRecipesList = buildRecipeList(recipeData.map(recipe => recipe.ID));
    siteData.allRecipes.listGroup = allRecipesList;

    // accordion stuff
    attachToAccordion(courseParentAccordion, courseChildAccordion);
    attachToAccordion(cuisinesParentAccordion, cuisinesChildAccordion);
    attachToAccordion(allRecipesAccordion, allRecipesList);
    applySort();
    writeRecipeCounts();

    // header controls
    const dimen = "48"; // (double the default icon size)
    feather.replace({ height: dimen, width: dimen});
    attachIconControls();
}

function normalizeString(str) {
    const latinized = str.normalize("NFD").replace(/[\u0300-\u036f]/g, "");
    const alphaNumeric = latinized.replaceAll(nonAlphaNumericRegexp, "");
    return alphaNumeric.toLowerCase();
}

function augmentRecipeData(recipeData) {
    return recipeData.map(recipe => {
        const clonedRecipe = {
            ...recipe
        };
        clonedRecipe.cleanName = normalizeString(recipe.Name);
        clonedRecipe.Keywords = recipe.Keywords.map(normalizeString);
        clonedRecipe.Ingredients = recipe.Keywords.map(normalizeString);
        clonedRecipe.isVisible = true;
        return clonedRecipe;
    });
}

function initialize() {
    annotatedRecipeData = augmentRecipeData(recipeData);
    iframe.src = "";
    siteData = {
        courses: {
            listGroups: [],
            recipes: []
        },
        cuisines: {
            listGroups: [],
            recipes: []
        },
        allRecipes: {
            listGroup: null,
            recipes: []
        },
        isDarkMode: false,
        searchTerm: ""
    }
    indexRecipes(annotatedRecipeData);
    buildSite();
}

initialize();