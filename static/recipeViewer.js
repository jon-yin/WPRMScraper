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

function buildHeader() {
    const menuBar = document.querySelector("#menubar");
    const dimen = menuBar.offsetHeight;
    const iconEnd = createElement({
        type: "span",
        classes: ["icon-end"]
    });
    const iconStart = createElement({
        type: "span",
        classes: ["icon-start"]
    })

    const createIconStart = (svg, tooltip = "") => {
        const cloned = iconStart.cloneNode();
        cloned.innerHTML = svg;
        if (tooltip !== "") {
            cloned.setAttribute("data-bs-toggle", "tooltip");
            cloned.setAttribute("data-bs-title", tooltip);
            cloned.setAttribute("data-bs-placement", "bottom");
            cloned.tooltip = new bootstrap.Tooltip(cloned);
        }
        return cloned;
    }

    const createIconEnd = (svg, tooltip = "") => {
        const cloned = iconEnd.cloneNode();
        cloned.innerHTML = svg;
        if (tooltip !== "") {
            cloned.setAttribute("data-bs-toggle", "tooltip");
            cloned.setAttribute("data-bs-title", tooltip);
            cloned.setAttribute("data-bs-placement", "bottom");
            cloned.tooltip = new bootstrap.Tooltip(cloned);
        }
        return cloned;
    }
    
    const sun = feather.icons["sun"].toSvg({ height: dimen, width: dimen });
    const moon = feather.icons["moon"].toSvg({ height: dimen, width: dimen });
    const expand = feather.icons["maximize-2"].toSvg({height: dimen, width: dimen});
    const collapse = feather.icons["minimize-2"].toSvg({height: dimen, width: dimen});
    
    const sunElem = createIconEnd(sun, "light mode");
    const moonElem = createIconEnd(moon, "dark mode");
    const expandElem = createIconStart(expand, "expand all recipes");
    const collapseElem = createIconStart(collapse, "collapse all recipes");

    // Nightmode toggle
    const toggleNightMode = (e) => {
        siteData.isNightMode = !siteData.isNightMode;
        e.currentTarget.tooltip.hide();
        if (siteData.isNightMode) {
            e.currentTarget.replaceWith(sunElem);
            document.querySelector("html").setAttribute("data-bs-theme", "dark");
        } else {
            e.currentTarget.replaceWith(moonElem);
            document.querySelector("html").setAttribute("data-bs-theme", "light");
        }

    } 
    sunElem.addEventListener("click", toggleNightMode);
    moonElem.addEventListener("click", toggleNightMode);

    // expand and collapse controls
    expandElem.addEventListener("click", () => {
        const divs = document.querySelectorAll("div.collapse");
        divs.forEach(div => div.classList.add("show"));
    });

    collapseElem.addEventListener("click", () => {
        const divs = document.querySelectorAll("div.collapse");
        divs.forEach(div => div.classList.remove("show"));
    });

    // Search field
    const searchField = document.querySelector("#template").content.querySelector("#searchField");
    const clearIcon = feather.icons["x"].toSvg({height: dimen, width: dimen});
    const clearElem = createIconStart(clearIcon, "clear search");
    clearElem.id = "clear";
    searchField.append(clearElem);
    clearElem.addEventListener("click", e => {
        document.querySelector("#search").value = "";
    });

    const nightModeElem = siteData.isNightMode ? sunElem : moonElem;

    menuBar.append(expandElem, collapseElem, searchField, nightModeElem);
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
                            children: [key.charAt(0).toUpperCase() + key.slice(1)],
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

    // header controls
    buildHeader();
}

function initialize() {
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
        isNightMode: false
    }
    indexRecipes(recipeData);
    buildSite();
}

initialize();