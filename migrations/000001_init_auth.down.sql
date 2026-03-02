DROP TABLE IF EXISTS withdrawals;
DROP TABLE IF EXISTS orders;

-- we have to remove 'users' last because other tables have foreign key to 'users'
DROP TABLE IF EXISTS users;

-- index will be removed automatically after removing tables
