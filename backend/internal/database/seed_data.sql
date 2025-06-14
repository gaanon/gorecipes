-- Seed data for GoRecipes PostgreSQL database
-- This file contains sample data for testing and development

-- Insert sample ingredients
INSERT INTO ingredients (id, name) VALUES 
    (uuid_generate_v4(), 'flour'),
    (uuid_generate_v4(), 'eggs'),
    (uuid_generate_v4(), 'milk'),
    (uuid_generate_v4(), 'butter'),
    (uuid_generate_v4(), 'sugar'),
    (uuid_generate_v4(), 'salt'),
    (uuid_generate_v4(), 'baking powder'),
    (uuid_generate_v4(), 'vanilla extract'),
    (uuid_generate_v4(), 'tomatoes'),
    (uuid_generate_v4(), 'onions'),
    (uuid_generate_v4(), 'garlic'),
    (uuid_generate_v4(), 'olive oil'),
    (uuid_generate_v4(), 'basil'),
    (uuid_generate_v4(), 'pasta'),
    (uuid_generate_v4(), 'cheese')
ON CONFLICT (name) DO NOTHING;

-- Insert sample recipes
INSERT INTO recipes (id, name, method, photo_filename, created_at, updated_at) VALUES 
    (
        uuid_generate_v4(), 
        'Classic Pancakes',
        E'1. Mix dry ingredients in a large bowl\n2. In another bowl, whisk together milk, eggs, and melted butter\n3. Combine wet and dry ingredients until just mixed\n4. Cook on a hot griddle until bubbles form and edges are set\n5. Flip and cook until golden brown',
        'placeholder.jpg',
        NOW() - INTERVAL '2 days',
        NOW() - INTERVAL '2 days'
    ),
    (
        uuid_generate_v4(),
        'Simple Pasta Marinara',
        E'1. Heat olive oil in a large pan\n2. Sauté diced onions until translucent\n3. Add minced garlic and cook for 1 minute\n4. Add tomatoes, salt, and basil\n5. Simmer for 20 minutes\n6. Serve over cooked pasta with cheese',
        'placeholder.jpg',
        NOW() - INTERVAL '1 day',
        NOW() - INTERVAL '1 day'
    )
ON CONFLICT DO NOTHING;

-- Link recipes to ingredients
-- Note: Using subqueries to link by name, assuming names are unique as per ON CONFLICT clauses.

-- For the Classic Pancakes recipe

-- Sample meal plan entries (for testing)
INSERT INTO meal_plan_entries (recipe_id, date)
SELECT id, CURRENT_DATE + INTERVAL '1 day'
FROM recipes WHERE name = 'Classic Pancakes'
LIMIT 1
ON CONFLICT DO NOTHING;

INSERT INTO meal_plan_entries (recipe_id, date)
SELECT id, CURRENT_DATE + INTERVAL '2 days'
FROM recipes WHERE name = 'Simple Pasta Marinara'
LIMIT 1
ON CONFLICT DO NOTHING;

-- Update comments for clarity
COMMENT ON TABLE ingredients IS 'Sample ingredients for testing and development.';
COMMENT ON TABLE recipes IS 'Sample recipes for testing and development.';
COMMENT ON TABLE recipe_ingredients IS 'Links recipes to their ingredients with quantities.';
COMMENT ON TABLE meal_plan_entries IS 'Sample meal plan entries for testing.';

