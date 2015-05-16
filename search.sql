-- phpMyAdmin SQL Dump
-- version 4.4.3
-- http://www.phpmyadmin.net
--
-- Host: localhost
-- Erstellungszeit: 16. Mai 2015 um 14:51
-- Server-Version: 5.6.24
-- PHP-Version: 5.6.8

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;

--
-- Datenbank: `search`
--

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `crawl`
--

CREATE TABLE IF NOT EXISTS `crawl` (
  `url` varchar(255) NOT NULL,
  `timestamp` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `keyword_url`
--

CREATE TABLE IF NOT EXISTS `keyword_url` (
  `fk_keyword` varchar(45) NOT NULL DEFAULT '',
  `fk_url` varchar(255) NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Indizes der exportierten Tabellen
--

--
-- Indizes für die Tabelle `crawl`
--
ALTER TABLE `crawl`
  ADD PRIMARY KEY (`url`),
  ADD UNIQUE KEY `url_UNIQUE` (`url`);

--
-- Indizes für die Tabelle `keyword_url`
--
ALTER TABLE `keyword_url`
  ADD PRIMARY KEY (`fk_keyword`,`fk_url`),
  ADD KEY `fk_keyword_idx` (`fk_keyword`);

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
