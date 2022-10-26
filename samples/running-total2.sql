-- A version of running-total1 that adds some comments 
SELECT DATE_FORMAT(co.order_date, '%Y-%m') AS order_month,
       DATE_FORMAT(co.order_date, '%Y-%m-%d') AS order_day,
       COUNT(DISTINCT co.order_id) AS num_orders,
       COUNT(ol.book_id) AS num_books, -- Number of books
       SUM(ol.price) AS total_price,   -- Their price total
       SUM(COUNT(ol.book_id)) OVER (
          PARTITION BY DATE_FORMAT(co.order_date, '%Y-%m')
          ORDER BY DATE_FORMAT(co.order_date, '%Y-%m-%d')
        ) AS running_total_num_books  -- This magic keeps a running total
FROM cust_order co
INNER JOIN order_line ol 
ON co.order_id = ol.order_id
GROUP BY DATE_FORMAT(co.order_date, '%Y-%m'),   -- Group by mont
         DATE_FORMAT(co.order_date, '%Y-%m-%d') -- Secondarily by day
ORDER BY co.order_date ASC;  -- Make sure we keep date order in report
