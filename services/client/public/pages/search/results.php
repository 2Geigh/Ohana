<?php if ($err !== null): ?>
    <h1>Hmm... looks like we've encountered an
        error on our part.</h1>
    <h2>Feel free to read through it if you're a nerd like us:</h2>
    <p><?php echo "{$err}" ?></p>
<?php else: ?>
    <h1>Results for <?php echo "{$query}" ?></h1>
<?php endif ?>