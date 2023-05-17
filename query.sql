SELECT name AS spot_name,
       CASE
           WHEN website ~ '^https?://([^/]+)' THEN substring(website from '^https?://([^/]+)')
           ELSE website
       END AS domain,
       COUNT(*) AS domain_count
FROM "MY_TABLE"
WHERE website IS NOT NULL AND website != ''
GROUP BY name, domain
HAVING COUNT(*) > 1;
