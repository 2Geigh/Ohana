<?php
require_once PROJECT_ROOT . '/src/util/Language.php';
?>

<?php if ($err !== null): ?>
    <h1>Hmm... looks like we've encountered an
        error on our part.</h1>
    <h2>Feel free to read through it if you're a nerd like us:</h2>
    <p><?= $err ?></p>
<?php else: ?>
    <h1>Found <?= sizeof($results->Results) ?>
        results for <i>
            <?= $results->InputtedQuery ?>
        </i>
        in <?= $results->GetDuration_s() ?> seconds:
    </h1>
    <?php foreach ($results->Results as $result): ?>
        <li>
            <a href="<?= htmlspecialchars((string) $result->Url, ENT_QUOTES, 'UTF-8') ?>">
                [<?php
                $languageCode = htmlspecialchars((string) $result->Language, ENT_QUOTES, 'UTF-8');

                if (key_exists($languageCode, $LanguageEmojis)) {
                    echo $LanguageEmojis[$languageCode];
                } else {
                    echo $LanguageEmojis['xx'];
                }

                ?>]
                <?= htmlspecialchars((string) $result->Title, ENT_QUOTES, 'UTF-8') ?>
            </a>
            <p>
                <?= htmlspecialchars((string) $result->Description, ENT_QUOTES, 'UTF-8') ?>
            </p>
        </li>
    <?php endforeach; ?>
<?php endif ?>