<?php if ($err !== null): ?>
    <h1>Hmm... looks like we've encountered an
        error on our part.</h1>
    <h2>Feel free to read through it if you're a nerd like us:</h2>
    <p><?= $err ?></p>
<?php else: ?>
    <h1><?= sizeof($results) ?> results for <?= $query ?>:</h1>
    <?php foreach ($results as $key => $value): ?>
        <li>
            <?= htmlspecialchars((string) $value, ENT_QUOTES, 'UTF-8') ?>
        </li>
    <?php endforeach; ?>
<?php endif ?>