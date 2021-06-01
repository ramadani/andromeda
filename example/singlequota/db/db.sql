CREATE TABLE `histories` (
  `id` varchar(36) NOT NULL,
  `voucher_id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE `vouchers` (
  `id` varchar(36) NOT NULL,
  `code` varchar(25) NOT NULL,
  `quota_limit` int(11) NOT NULL,
  `quota_usage` int(11) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

INSERT INTO `vouchers` (`id`, `code`, `quota_limit`, `quota_usage`) VALUES
('5ab06f8f-07c3-4de4-bb3f-8a3530dd0fd5', 'VC456', 300, 0),
('7050642c-94f5-47e4-bb5a-0b98ff77acc7', 'VC123', 200, 0);